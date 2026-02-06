package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/docuseal/docuseal-cli/internal/validation"
	"github.com/spf13/cobra"
)

var submittersCmd = &cobra.Command{
	Use:     "submitters",
	Aliases: []string{"submitter", "signers", "signer"},
	Short:   "Manage submitters",
	Long:    `List, view, and update submission submitters.`,
}

var submittersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List submitters",
	Long:  `List submitters with optional filtering by submission.`,
	RunE:  runSubmittersList,
}

var submittersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get submitter details",
	Long:  `Retrieve detailed information about a specific submitter.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSubmittersGet,
}

var submittersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update submitter",
	Long: `Update a submitter's details or mark as completed.

Use --completed to programmatically complete signing on behalf of the submitter.
This is useful for auto-signing workflows.`,
	Example: `  # Update submitter email
  docuseal submitters update 456 --email newemail@example.com

  # Mark as completed (auto-sign)
  docuseal submitters update 456 --completed

  # Send notification with custom message
  docuseal submitters update 456 --send-email --message-subject "Please sign" --message-body "Sign here"

  # Pre-fill field values
  docuseal submitters update 456 --values '{"field_name":"value"}'

  # Set custom metadata
  docuseal submitters update 456 --metadata '{"user_id":"123"}'

  # Configure fields with defaults and validation
  docuseal submitters update 456 --fields '[{"name":"field1","default_value":"default","readonly":true}]'

  # Require phone 2FA and set redirect
  docuseal submitters update 456 --require-phone-2fa --completed-redirect-url https://example.com/thanks`,
	Args: cobra.ExactArgs(1),
	RunE: runSubmittersUpdate,
}

// Flags
var (
	submittersLimit             int
	submittersSubmissionID      int
	submittersEmail             string
	submittersName              string
	submittersPhone             string
	submittersCompleted         bool
	submittersSendEmail         bool
	submittersSendSMS           bool
	submittersValues            string
	submittersExternalID        string
	submittersReplyTo           string
	submittersMetadata          string
	submittersCompletedRedirect string
	submittersRequirePhone2FA   bool
	submittersFields            string
	submittersMessageSubject    string
	submittersMessageBody       string
)

func init() {
	rootCmd.AddCommand(submittersCmd)

	submittersCmd.AddCommand(submittersListCmd)
	submittersCmd.AddCommand(submittersGetCmd)
	submittersCmd.AddCommand(submittersUpdateCmd)

	// List flags
	submittersListCmd.Flags().IntVar(&submittersLimit, "limit", 0, "Maximum number of submitters to return")
	submittersListCmd.Flags().IntVar(&submittersSubmissionID, "submission-id", 0, "Filter by submission ID")

	// Update flags
	submittersUpdateCmd.Flags().StringVar(&submittersEmail, "email", "", "New email address")
	submittersUpdateCmd.Flags().StringVar(&submittersName, "name", "", "New name")
	submittersUpdateCmd.Flags().StringVar(&submittersPhone, "phone", "", "New phone number")
	submittersUpdateCmd.Flags().BoolVar(&submittersCompleted, "completed", false, "Mark as completed (auto-sign)")
	submittersUpdateCmd.Flags().BoolVar(&submittersSendEmail, "send-email", false, "Send notification email")
	submittersUpdateCmd.Flags().BoolVar(&submittersSendSMS, "send-sms", false, "Send notification SMS")
	submittersUpdateCmd.Flags().StringVar(&submittersValues, "values", "", "Pre-fill field values (JSON string, e.g., '{\"field_name\":\"value\"}')")
	submittersUpdateCmd.Flags().StringVar(&submittersExternalID, "external-id", "", "App-specific identifier")
	submittersUpdateCmd.Flags().StringVar(&submittersReplyTo, "reply-to", "", "Reply-To address for emails")
	submittersUpdateCmd.Flags().StringVar(&submittersMetadata, "metadata", "", "Custom metadata (JSON string)")
	submittersUpdateCmd.Flags().StringVar(&submittersCompletedRedirect, "completed-redirect-url", "", "Redirect URL after completion")
	submittersUpdateCmd.Flags().BoolVar(&submittersRequirePhone2FA, "require-phone-2fa", false, "Require phone verification")
	submittersUpdateCmd.Flags().StringVar(&submittersFields, "fields", "", "Field configurations (JSON string)")
	submittersUpdateCmd.Flags().StringVar(&submittersMessageSubject, "message-subject", "", "Custom email subject")
	submittersUpdateCmd.Flags().StringVar(&submittersMessageBody, "message-body", "", "Custom email body")
}

func runSubmittersList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	limit := submittersLimit
	reqLimit := limit
	if ((mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta)) && limit > 0 {
		reqLimit = limit + 1
	}

	submitters, err := client.ListSubmitters(cmd.Context(), reqLimit, submittersSubmissionID)
	if err != nil {
		return fmt.Errorf("failed to list submitters: %w", err)
	}

	if (mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta) {
		out := submitters
		hasMore := false
		if limit > 0 && len(out) > limit {
			hasMore = true
			out = out[:limit]
		}

		if mode == outfmt.JSON && !bareJSON {
			env := makeListEnvelope(out, len(out), limit, 0, 0, hasMore, 0, 0)
			env["submission_id"] = submittersSubmissionID
			outputResult(mode, env, func() {})
			return nil
		}

		meta := map[string]any{
			"_meta": map[string]any{
				"count":         len(out),
				"limit":         limit,
				"has_more":      hasMore,
				"submission_id": submittersSubmissionID,
			},
		}
		stream := make([]any, 0, len(out)+1)
		for _, s := range out {
			stream = append(stream, s)
		}
		stream = append(stream, meta)
		outputResult(mode, stream, func() {})
		return nil
	}

	outputResult(mode, submitters, func() {
		if len(submitters) == 0 {
			fmt.Println("No submitters found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "ID\tEMAIL\tROLE\tSTATUS\tSUBMISSION"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, s := range submitters {
			if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n",
				s.ID,
				truncateString(s.Email, 30),
				s.Role,
				s.Status,
				s.SubmissionID,
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}

		// Pagination hint
		if submittersLimit > 0 && len(submitters) == submittersLimit {
			fmt.Fprintf(os.Stderr, "\n# More results may be available. Use --limit with higher value.\n")
		}
	})

	return nil
}

func runSubmittersGet(cmd *cobra.Command, args []string) error {
	id, err := parseIDArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid submitter ID: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	submitter, err := client.GetSubmitter(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get submitter: %w", err)
	}

	outputResult(mode, submitter, func() {
		fmt.Printf("ID: %d\n", submitter.ID)
		fmt.Printf("Email: %s\n", submitter.Email)
		if submitter.Name != "" {
			fmt.Printf("Name: %s\n", submitter.Name)
		}
		if submitter.Phone != "" {
			fmt.Printf("Phone: %s\n", submitter.Phone)
		}
		fmt.Printf("Role: %s\n", submitter.Role)
		fmt.Printf("Status: %s\n", submitter.Status)
		fmt.Printf("Submission ID: %d\n", submitter.SubmissionID)
		fmt.Printf("Created: %s\n", formatTime(submitter.CreatedAt))
		if submitter.SentAt != nil {
			fmt.Printf("Sent: %s\n", formatTimePtr(submitter.SentAt))
		}
		if submitter.OpenedAt != nil {
			fmt.Printf("Opened: %s\n", formatTimePtr(submitter.OpenedAt))
		}
		if submitter.CompletedAt != nil {
			fmt.Printf("Completed: %s\n", formatTimePtr(submitter.CompletedAt))
		}
		if submitter.DeclinedAt != nil {
			fmt.Printf("Declined: %s\n", formatTimePtr(submitter.DeclinedAt))
		}
		if submitter.EmbedSrc != "" {
			fmt.Printf("Sign URL: %s\n", submitter.EmbedSrc)
		}
		if len(submitter.Values) > 0 {
			fmt.Println("Values:")
			for _, v := range submitter.Values {
				fmt.Printf("  %s: %v\n", v.Field, v.Value)
			}
		}
	})

	return nil
}

func runSubmittersUpdate(cmd *cobra.Command, args []string) error {
	id, err := parseIDArg(args[0])
	if err != nil {
		return fmt.Errorf("invalid submitter ID: %w", err)
	}

	// Validate email addresses if provided
	if submittersEmail != "" {
		if err := validation.ValidateEmail(submittersEmail); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
	}

	if submittersReplyTo != "" {
		if err := validation.ValidateEmail(submittersReplyTo); err != nil {
			return fmt.Errorf("invalid reply-to email: %w", err)
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	req := &api.UpdateSubmitterRequest{
		Email:                submittersEmail,
		Name:                 submittersName,
		Phone:                submittersPhone,
		Completed:            submittersCompleted,
		SendEmail:            submittersSendEmail,
		SendSMS:              submittersSendSMS,
		ExternalID:           submittersExternalID,
		ReplyTo:              submittersReplyTo,
		CompletedRedirectURL: submittersCompletedRedirect,
		RequirePhone2FA:      submittersRequirePhone2FA,
	}

	// Parse JSON fields
	if submittersValues != "" {
		var values map[string]any
		if err := json.Unmarshal([]byte(submittersValues), &values); err != nil {
			return fmt.Errorf("invalid values JSON: %w", err)
		}
		req.Values = values
	}

	if submittersMetadata != "" {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(submittersMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		req.Metadata = metadata
	}

	if submittersFields != "" {
		var fields []api.FieldConfig
		if err := json.Unmarshal([]byte(submittersFields), &fields); err != nil {
			return fmt.Errorf("invalid fields JSON: %w", err)
		}
		req.Fields = fields
	}

	// Handle custom message
	if submittersMessageSubject != "" || submittersMessageBody != "" {
		req.Message = &api.Message{
			Subject: submittersMessageSubject,
			Body:    submittersMessageBody,
		}
	}

	submitter, err := client.UpdateSubmitter(cmd.Context(), id, req)
	if err != nil {
		return fmt.Errorf("failed to update submitter: %w", err)
	}

	outputResult(mode, submitter, func() {
		fmt.Printf("Updated submitter %d\n", submitter.ID)
		fmt.Printf("Status: %s\n", submitter.Status)
		if submittersCompleted {
			fmt.Println("Marked as completed")
		}
	})

	return nil
}
