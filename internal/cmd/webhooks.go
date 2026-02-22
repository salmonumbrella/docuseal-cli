package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:     "webhooks",
	Aliases: []string{"webhook", "wh"},
	Short:   "Manage webhooks",
	Long:    `List, create, update, and manage DocuSeal webhooks.`,
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	Long:  `List all configured webhooks.`,
	Example: `  # List all webhooks
  docuseal webhooks list

  # List with JSON output
  docuseal webhooks list -o json

  # List with pagination
  docuseal webhooks list --limit 10 --after 5`,
	RunE: runWebhooksList,
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get webhook details",
	Long:  `Retrieve detailed information about a specific webhook.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWebhooksGet,
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create webhook",
	Long:  `Create a new webhook to receive event notifications.`,
	Example: `  # Create webhook for submission completed events
  docuseal webhooks create --url https://example.com/webhook --events submission.completed

  # Create webhook for multiple events
  docuseal webhooks create --url https://example.com/webhook \
    --events submission.created \
    --events submission.completed \
    --events submission.archived`,
	RunE: runWebhooksCreate,
}

var webhooksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update webhook",
	Long:  `Update an existing webhook's URL, events, or active status.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Update webhook URL
  docuseal webhooks update 123 --url https://example.com/new-webhook

  # Update events
  docuseal webhooks update 123 --events submission.completed --events submission.archived

  # Disable webhook
  docuseal webhooks update 123 --active false

  # Enable webhook
  docuseal webhooks update 123 --active true`,
	RunE: runWebhooksUpdate,
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete webhook",
	Long:  `Delete a webhook configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWebhooksDelete,
}

// Flags
var (
	webhooksLimit     int
	webhooksAfter     int
	webhooksBefore    int
	webhooksURL       string
	webhooksEvents    []string
	webhooksActive    string
	webhookShowSecret bool
)

// maskSecret masks a secret string by showing only first and last 4 characters
func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

func init() {
	rootCmd.AddCommand(webhooksCmd)

	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksGetCmd)
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCmd.AddCommand(webhooksUpdateCmd)
	webhooksCmd.AddCommand(webhooksDeleteCmd)

	// List flags
	webhooksListCmd.Flags().IntVar(&webhooksLimit, "limit", 0, "Maximum number of webhooks to return")
	webhooksListCmd.Flags().IntVar(&webhooksAfter, "after", 0, "Pagination cursor, get IDs greater than value")
	webhooksListCmd.Flags().IntVar(&webhooksBefore, "before", 0, "Pagination cursor, get IDs less than value")

	// Create flags
	webhooksCreateCmd.Flags().StringVar(&webhooksURL, "url", "", "Webhook URL (required)")
	webhooksCreateCmd.Flags().StringArrayVar(&webhooksEvents, "events", []string{}, "Event types to subscribe to (required, can be specified multiple times)")
	webhooksCreateCmd.Flags().BoolVar(&webhookShowSecret, "show-secret", false, "Show webhook secret in output (security risk)")
	mustMarkFlagRequired(webhooksCreateCmd, "url")
	mustMarkFlagRequired(webhooksCreateCmd, "events")

	// Get flags
	webhooksGetCmd.Flags().BoolVar(&webhookShowSecret, "show-secret", false, "Show webhook secret in output (security risk)")

	// Update flags
	webhooksUpdateCmd.Flags().StringVar(&webhooksURL, "url", "", "New webhook URL")
	webhooksUpdateCmd.Flags().StringArrayVar(&webhooksEvents, "events", []string{}, "Event types to subscribe to (can be specified multiple times)")
	webhooksUpdateCmd.Flags().StringVar(&webhooksActive, "active", "", "Enable or disable webhook (true/false)")
}

func runWebhooksList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	limit := webhooksLimit
	reqLimit := limit
	if ((mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta)) && limit > 0 {
		reqLimit = limit + 1
	}

	webhooks, err := client.ListWebhooks(cmd.Context(), reqLimit, webhooksAfter, webhooksBefore)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if (mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta) {
		out := webhooks
		hasMore := false
		if limit > 0 && len(out) > limit {
			hasMore = true
			out = out[:limit]
		}
		nextAfter := 0
		nextBefore := 0
		if len(out) > 0 {
			nextBefore = out[0].ID
			if hasMore {
				nextAfter = out[len(out)-1].ID
			}
		}

		if mode == outfmt.JSON && !bareJSON {
			env := makeListEnvelope(out, len(out), limit, webhooksAfter, webhooksBefore, hasMore, nextAfter, nextBefore)
			outputResult(mode, env, func() {})
			return nil
		}

		meta := map[string]any{
			"_meta": map[string]any{
				"count":       len(out),
				"limit":       limit,
				"after":       webhooksAfter,
				"before":      webhooksBefore,
				"has_more":    hasMore,
				"next_after":  nextAfter,
				"next_before": nextBefore,
			},
		}
		stream := make([]any, 0, len(out)+1)
		for _, wh := range out {
			stream = append(stream, wh)
		}
		stream = append(stream, meta)
		outputResult(mode, stream, func() {})
		return nil
	}

	outputResult(mode, webhooks, func() {
		if len(webhooks) == 0 {
			fmt.Println("No webhooks found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "ID\tURL\tEVENTS\tACTIVE\tCREATED"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, wh := range webhooks {
			events := strings.Join(wh.Events, ", ")
			active := "Yes"
			if !wh.Active {
				active = "No"
			}
			if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				wh.ID,
				truncateString(wh.URL, 50),
				truncateString(events, 40),
				active,
				formatTime(wh.CreatedAt),
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}

		// Pagination hint
		if webhooksLimit > 0 && len(webhooks) == webhooksLimit {
			fmt.Fprintf(os.Stderr, "\n# More results may be available. Use --after %d to see next page.\n", webhooks[len(webhooks)-1].ID)
		}
	})

	return nil
}

func runWebhooksGet(cmd *cobra.Command, args []string) error {
	id, err := parseIDArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	webhook, err := client.GetWebhook(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	outputResult(mode, webhook, func() {
		fmt.Printf("ID: %d\n", webhook.ID)
		fmt.Printf("URL: %s\n", webhook.URL)
		fmt.Printf("Events: %s\n", strings.Join(webhook.Events, ", "))
		fmt.Printf("Active: %t\n", webhook.Active)
		if webhook.Secret != "" {
			if webhookShowSecret {
				fmt.Printf("Secret: %s\n", webhook.Secret)
			} else {
				fmt.Printf("Secret: %s (use --show-secret to reveal)\n", maskSecret(webhook.Secret))
			}
		}
		fmt.Printf("Created: %s\n", formatTime(webhook.CreatedAt))
		fmt.Printf("Updated: %s\n", formatTime(webhook.UpdatedAt))
	})

	return nil
}

func runWebhooksCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	req := &api.CreateWebhookRequest{
		URL:    webhooksURL,
		Events: webhooksEvents,
	}

	webhook, err := client.CreateWebhook(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	outputResult(mode, webhook, func() {
		fmt.Printf("Created webhook %d\n", webhook.ID)
		fmt.Printf("URL: %s\n", webhook.URL)
		fmt.Printf("Events: %s\n", strings.Join(webhook.Events, ", "))
		if webhook.Secret != "" {
			if webhookShowSecret {
				fmt.Printf("Secret: %s\n", webhook.Secret)
			} else {
				fmt.Printf("Secret: %s (use --show-secret to reveal)\n", maskSecret(webhook.Secret))
			}
		}
	})

	return nil
}

func runWebhooksUpdate(cmd *cobra.Command, args []string) error {
	id, err := parseIDArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	req := &api.UpdateWebhookRequest{
		URL:    webhooksURL,
		Events: webhooksEvents,
	}

	// Parse active flag if provided
	if webhooksActive != "" {
		active, err := strconv.ParseBool(webhooksActive)
		if err != nil {
			return fmt.Errorf("invalid active value: must be true or false")
		}
		req.Active = &active
	}

	webhook, err := client.UpdateWebhook(cmd.Context(), id, req)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	outputResult(mode, webhook, func() {
		fmt.Printf("Updated webhook %d\n", webhook.ID)
		fmt.Printf("URL: %s\n", webhook.URL)
		fmt.Printf("Events: %s\n", strings.Join(webhook.Events, ", "))
		fmt.Printf("Active: %t\n", webhook.Active)
	})

	return nil
}

func runWebhooksDelete(cmd *cobra.Command, args []string) error {
	id, err := parseIDArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid webhook ID: %w", err)
	}

	if dryRunPreview("delete webhook %d", id) {
		return nil
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	err = client.DeleteWebhook(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("Deleted webhook %d\n", id)
	return nil
}
