package cmd

import (
	"fmt"
	"os"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View events and webhooks",
	Long:  `List form and submission events.`,
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	Long: `List events by type.

Event types for forms: view, start, complete
Event types for submissions: created, completed, archived`,
	RunE: runEventsList,
}

var (
	eventsType         string
	eventsCategory     string
	eventsSubmissionID int
	eventsLimit        int
)

func init() {
	rootCmd.AddCommand(eventsCmd)
	eventsCmd.AddCommand(eventsListCmd)

	eventsListCmd.Flags().StringVar(&eventsCategory, "category", "submission", "Event category: form, submission")
	eventsListCmd.Flags().StringVar(&eventsType, "type", "completed", "Event type (e.g., view, start, complete, created, completed, archived)")
	eventsListCmd.Flags().IntVar(&eventsSubmissionID, "submission-id", 0, "Filter by submission ID (for submission events)")
	eventsListCmd.Flags().IntVar(&eventsLimit, "limit", 0, "Maximum number of events to return")
}

func runEventsList(cmd *cobra.Command, args []string) error {
	client, err := getClientOrError(cmd)
	if err != nil {
		return err
	}
	mode := getOutputMode()

	var events []api.Event
	var listErr error

	if eventsCategory == "form" {
		events, listErr = client.ListFormEvents(cmd.Context(), eventsType, eventsLimit)
	} else {
		events, listErr = client.ListSubmissionEvents(cmd.Context(), eventsType, eventsSubmissionID, eventsLimit)
	}

	if listErr != nil {
		return fmt.Errorf("failed to list events: %w", listErr)
	}

	outputResult(mode, events, func() {
		if len(events) == 0 {
			fmt.Println("No events found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "ID\tTYPE\tSUBMISSION_ID\tSUBMITTER_ID\tCREATED"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, e := range events {
			submissionID := "-"
			if e.SubmissionID > 0 {
				submissionID = fmt.Sprintf("%d", e.SubmissionID)
			}
			submitterID := "-"
			if e.SubmitterID > 0 {
				submitterID = fmt.Sprintf("%d", e.SubmitterID)
			}
			if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				e.ID,
				e.EventType,
				submissionID,
				submitterID,
				formatTime(e.CreatedAt),
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}

		// Pagination hint
		if eventsLimit > 0 && len(events) == eventsLimit {
			fmt.Fprintf(os.Stderr, "\n# More results may be available. Use --limit with higher value.\n")
		}
	})

	return nil
}
