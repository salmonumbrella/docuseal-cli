package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/docuseal/docuseal-cli/internal/validation"
	"github.com/spf13/cobra"
)

var submissionsCmd = &cobra.Command{
	Use:     "submissions",
	Aliases: []string{"submission", "sub", "s"},
	Short:   "Manage submissions",
	Long:    `List, create, and manage document signing submissions.`,
}

var submissionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List submissions",
	Long:  `List submissions with optional filtering by template, status, query, and more.`,
	Example: `  # List all submissions
  docuseal submissions list

  # Filter by template
  docuseal submissions list --template-id 123

  # Filter by status
  docuseal submissions list --status completed

  # Search by submitter name/email/phone
  docuseal submissions list --query "john@example.com"
  docuseal submissions list -q "John Doe"

  # Filter by submission slug
  docuseal submissions list --slug "contract-abc123"

  # Filter by template folder
  docuseal submissions list --template-folder "HR Documents"

  # Show archived submissions
  docuseal submissions list --archived

  # Pagination: get submissions with IDs greater than 100
  docuseal submissions list --after 100 --limit 50

  # Pagination: get submissions with IDs less than 500
  docuseal submissions list --before 500 --limit 50

  # Combine multiple filters
  docuseal submissions list --template-id 123 --status pending --query "john"`,
	RunE: runSubmissionsList,
}

var submissionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get submission details",
	Long:  `Retrieve detailed information about a specific submission.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSubmissionsGet,
}

var submissionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create submission from template",
	Long:  `Create a new submission from an existing template.`,
	Example: `  # Create submission with single signer
  docuseal submissions create --template-id 123 --submitters "john@example.com:Signer"

  # Create with multiple signers
  docuseal submissions create --template-id 123 \
    --submitters "john@example.com:Signer" \
    --submitters "jane@example.com:Approver" \
    --send-email

  # Create with custom message
  docuseal submissions create --template-id 123 \
    --submitters "john@example.com:Signer" \
    --message "Please Sign:Please review and sign this document"

  # Create with SMS notification
  docuseal submissions create --template-id 123 \
    --submitters "john@example.com:Signer" \
    --send-sms

  # Create with completion redirect and BCC
  docuseal submissions create --template-id 123 \
    --submitters "john@example.com:Signer" \
    --completed-redirect-url "https://example.com/done" \
    --bcc-completed "admin@example.com"

  # Create with expiration
  docuseal submissions create --template-id 123 \
    --submitters "john@example.com:Signer" \
    --expire-at "2025-12-31T23:59:59Z"`,
	RunE: runSubmissionsCreate,
}

var submissionsCreatePDFCmd = &cobra.Command{
	Use:   "create-pdf",
	Short: "Create submission from PDF",
	Long:  `Create a new submission directly from a PDF file.`,
	RunE:  runSubmissionsCreatePDF,
}

var submissionsCreateDOCXCmd = &cobra.Command{
	Use:   "create-docx",
	Short: "Create submission from DOCX",
	Long:  `Create a new submission directly from a DOCX file.`,
	RunE:  runSubmissionsCreateDOCX,
}

var submissionsCreateHTMLCmd = &cobra.Command{
	Use:   "create-html",
	Short: "Create submission from HTML",
	Long:  `Create a new submission directly from HTML content.`,
	RunE:  runSubmissionsCreateHTML,
}

var submissionsDocumentsCmd = &cobra.Command{
	Use:   "documents <id>",
	Short: "Get submission documents",
	Long:  `Retrieve the signed documents for a submission.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSubmissionsDocuments,
}

var submissionsArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive submission",
	Long:  `Archive a submission (soft delete).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSubmissionsArchive,
}

var submissionsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize submission without sending emails",
	Long: `Initialize a submission from a template without automatically sending emails.

Use this when you want to set up submitters and get signing URLs
without immediately notifying them via email.`,
	Example: `  # Initialize submission
  docuseal submissions init --template-id 123 --submitters "john@example.com:Signer"

  # Get signing URLs without sending emails
  docuseal submissions init --template-id 123 --submitters "john@example.com:Signer" -o json`,
	RunE: runSubmissionsInit,
}

var submissionsCreateEmailsCmd = &cobra.Command{
	Use:   "create-emails",
	Short: "Create submissions from comma-separated emails",
	Long: `Create submissions from a comma-separated list of emails.
This is a simplified endpoint for automation/Zapier workflows.`,
	Example: `  # Create submissions with email list
  docuseal submissions create-emails --template-id 123 --emails "a@example.com,b@example.com"

  # Create with email sending enabled
  docuseal submissions create-emails --template-id 123 --emails "a@example.com,b@example.com" --send-email

  # Create with custom message
  docuseal submissions create-emails --template-id 123 --emails "a@example.com,b@example.com" \
    --message-subject "Please sign" --message-body "Click link to sign"`,
	RunE: runSubmissionsCreateEmails,
}

// Flags
var (
	submissionsLimit                int
	submissionsTemplateID           int
	submissionsStatus               string
	submissionsQuery                string
	submissionsSlug                 string
	submissionsTemplateFolder       string
	submissionsArchived             bool
	submissionsAfter                int
	submissionsBefore               int
	submissionsSubmitters           []string
	submissionsSendEmail            bool
	submissionsSendSMS              bool
	submissionsMessage              string
	submissionsName                 string
	submissionsFile                 string
	submissionsHTML                 string
	submissionsVariables            []string
	submissionsEmails               string
	submissionsEmailsCSV            string
	submissionsMessageSubject       string
	submissionsMessageBody          string
	submissionsCompletedRedirectURL string
	submissionsBCCCompleted         string
	submissionsReplyTo              string
	submissionsExpireAt             string
)

func init() {
	rootCmd.AddCommand(submissionsCmd)

	submissionsCmd.AddCommand(submissionsListCmd)
	submissionsCmd.AddCommand(submissionsGetCmd)
	submissionsCmd.AddCommand(submissionsCreateCmd)
	submissionsCmd.AddCommand(submissionsCreatePDFCmd)
	submissionsCmd.AddCommand(submissionsCreateDOCXCmd)
	submissionsCmd.AddCommand(submissionsCreateHTMLCmd)
	submissionsCmd.AddCommand(submissionsDocumentsCmd)
	submissionsCmd.AddCommand(submissionsArchiveCmd)
	submissionsCmd.AddCommand(submissionsInitCmd)
	submissionsCmd.AddCommand(submissionsCreateEmailsCmd)

	// List flags
	submissionsListCmd.Flags().IntVar(&submissionsLimit, "limit", 0, "Maximum number of submissions to return")
	submissionsListCmd.Flags().IntVar(&submissionsTemplateID, "template-id", 0, "Filter by template ID")
	submissionsListCmd.Flags().StringVar(&submissionsStatus, "status", "", "Filter by status (pending, completed)")
	submissionsListCmd.Flags().StringVarP(&submissionsQuery, "query", "q", "", "Search by submitter name/email/phone")
	submissionsListCmd.Flags().StringVar(&submissionsSlug, "slug", "", "Filter by submission slug")
	submissionsListCmd.Flags().StringVar(&submissionsTemplateFolder, "template-folder", "", "Filter by template folder name")
	submissionsListCmd.Flags().BoolVar(&submissionsArchived, "archived", false, "Show archived submissions")
	submissionsListCmd.Flags().IntVar(&submissionsAfter, "after", 0, "Pagination cursor, get IDs greater than value")
	submissionsListCmd.Flags().IntVar(&submissionsBefore, "before", 0, "Pagination cursor, get IDs less than value")

	// Create flags
	submissionsCreateCmd.Flags().IntVar(&submissionsTemplateID, "template-id", 0, "Template ID (required)")
	submissionsCreateCmd.Flags().StringArrayVar(&submissionsSubmitters, "submitters", nil, "Submitters in EMAIL[:ROLE] format (can be repeated; ROLE optional when resolvable from template)")
	submissionsCreateCmd.Flags().StringVar(&submissionsEmailsCSV, "emails", "", "Comma-separated emails (shortcut; uses template roles/order)")
	submissionsCreateCmd.Flags().BoolVar(&submissionsSendEmail, "send-email", false, "Send email to submitters")
	submissionsCreateCmd.Flags().BoolVar(&submissionsSendSMS, "send-sms", false, "Send SMS notification to submitters")
	submissionsCreateCmd.Flags().StringVar(&submissionsMessage, "message", "", "Custom message in SUBJECT:BODY format")
	submissionsCreateCmd.Flags().StringVar(&submissionsCompletedRedirectURL, "completed-redirect-url", "", "URL to redirect after completion")
	submissionsCreateCmd.Flags().StringVar(&submissionsBCCCompleted, "bcc-completed", "", "BCC email address for completed documents")
	submissionsCreateCmd.Flags().StringVar(&submissionsReplyTo, "reply-to", "", "Reply-To address for notification emails")
	submissionsCreateCmd.Flags().StringVar(&submissionsExpireAt, "expire-at", "", "Expiration datetime (ISO 8601 format)")
	mustMarkFlagRequired(submissionsCreateCmd, "template-id")
	// Either --submitters or --emails must be provided (validated at runtime).

	// Init flags (reuse existing flags from create)
	submissionsInitCmd.Flags().IntVar(&submissionsTemplateID, "template-id", 0, "Template ID (required)")
	submissionsInitCmd.Flags().StringArrayVar(&submissionsSubmitters, "submitters", nil, "Submitters in EMAIL[:ROLE] format (required; ROLE optional when resolvable from template)")
	mustMarkFlagRequired(submissionsInitCmd, "template-id")
	mustMarkFlagRequired(submissionsInitCmd, "submitters")

	// Create emails flags
	submissionsCreateEmailsCmd.Flags().IntVar(&submissionsTemplateID, "template-id", 0, "Template ID (required)")
	submissionsCreateEmailsCmd.Flags().StringVar(&submissionsEmails, "emails", "", "Comma-separated list of emails (required)")
	submissionsCreateEmailsCmd.Flags().BoolVar(&submissionsSendEmail, "send-email", false, "Send email to submitters")
	submissionsCreateEmailsCmd.Flags().StringVar(&submissionsMessageSubject, "message-subject", "", "Custom email subject")
	submissionsCreateEmailsCmd.Flags().StringVar(&submissionsMessageBody, "message-body", "", "Custom email body")
	mustMarkFlagRequired(submissionsCreateEmailsCmd, "template-id")
	mustMarkFlagRequired(submissionsCreateEmailsCmd, "emails")

	// Create PDF flags
	submissionsCreatePDFCmd.Flags().StringVar(&submissionsFile, "file", "", "PDF file path (required)")
	submissionsCreatePDFCmd.Flags().StringArrayVar(&submissionsSubmitters, "submitters", nil, "Submitters in EMAIL[:ROLE] format (required; default ROLE: Signer)")
	submissionsCreatePDFCmd.Flags().StringVar(&submissionsName, "name", "", "Submission name")
	mustMarkFlagRequired(submissionsCreatePDFCmd, "file")
	mustMarkFlagRequired(submissionsCreatePDFCmd, "submitters")

	// Create DOCX flags
	submissionsCreateDOCXCmd.Flags().StringVar(&submissionsFile, "file", "", "DOCX file path (required)")
	submissionsCreateDOCXCmd.Flags().StringArrayVar(&submissionsSubmitters, "submitters", nil, "Submitters in EMAIL[:ROLE] format (required; default ROLE: Signer)")
	submissionsCreateDOCXCmd.Flags().StringVar(&submissionsName, "name", "", "Submission name")
	submissionsCreateDOCXCmd.Flags().StringArrayVar(&submissionsVariables, "variables", nil, "Variables in KEY=VALUE format")
	mustMarkFlagRequired(submissionsCreateDOCXCmd, "file")
	mustMarkFlagRequired(submissionsCreateDOCXCmd, "submitters")

	// Create HTML flags
	submissionsCreateHTMLCmd.Flags().StringVar(&submissionsHTML, "html", "", "HTML content (required)")
	submissionsCreateHTMLCmd.Flags().StringArrayVar(&submissionsSubmitters, "submitters", nil, "Submitters in EMAIL[:ROLE] format (required; default ROLE: Signer)")
	submissionsCreateHTMLCmd.Flags().StringVar(&submissionsName, "name", "", "Submission name")
	mustMarkFlagRequired(submissionsCreateHTMLCmd, "html")
	mustMarkFlagRequired(submissionsCreateHTMLCmd, "submitters")
}

func runSubmissionsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	limit := submissionsLimit
	reqLimit := limit
	if ((mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta)) && limit > 0 {
		reqLimit = limit + 1
	}

	submissions, err := client.ListSubmissions(
		cmd.Context(),
		reqLimit,
		submissionsTemplateID,
		submissionsStatus,
		submissionsQuery,
		submissionsSlug,
		submissionsTemplateFolder,
		submissionsArchived,
		submissionsAfter,
		submissionsBefore,
	)
	if err != nil {
		return fmt.Errorf("failed to list submissions: %w", err)
	}

	if (mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta) {
		out := submissions
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
			env := makeListEnvelope(out, len(out), limit, submissionsAfter, submissionsBefore, hasMore, nextAfter, nextBefore)
			outputResult(mode, env, func() {})
			return nil
		}

		meta := map[string]any{
			"_meta": map[string]any{
				"count":       len(out),
				"limit":       limit,
				"after":       submissionsAfter,
				"before":      submissionsBefore,
				"has_more":    hasMore,
				"next_after":  nextAfter,
				"next_before": nextBefore,
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

	outputResult(mode, submissions, func() {
		if len(submissions) == 0 {
			fmt.Println("No submissions found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "ID\tSTATUS\tTEMPLATE\tCREATED"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, s := range submissions {
			if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				s.ID,
				s.Status,
				truncateString(s.TemplateName, 30),
				formatTime(s.CreatedAt),
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}

		// Pagination hint
		if submissionsLimit > 0 && len(submissions) == submissionsLimit {
			fmt.Fprintf(os.Stderr, "\n# More results may be available. Use --after %d to see next page.\n", submissions[len(submissions)-1].ID)
		}
	})

	return nil
}

func runSubmissionsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveSubmissionID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	submission, err := client.GetSubmission(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	outputResult(mode, submission, func() {
		fmt.Printf("ID: %d\n", submission.ID)
		fmt.Printf("Status: %s\n", submission.Status)
		fmt.Printf("Slug: %s\n", submission.Slug)
		fmt.Printf("Template: %s (ID: %d)\n", submission.TemplateName, submission.TemplateID)
		fmt.Printf("Created: %s\n", formatTime(submission.CreatedAt))
		fmt.Printf("Updated: %s\n", formatTime(submission.UpdatedAt))
		if submission.CompletedAt != nil {
			fmt.Printf("Completed: %s\n", formatTimePtr(submission.CompletedAt))
		}
		if len(submission.Submitters) > 0 {
			fmt.Println("Submitters:")
			for _, sub := range submission.Submitters {
				fmt.Printf("  - %s (%s): %s\n", sub.Email, sub.Role, sub.Status)
			}
		}
	})

	return nil
}

func runSubmissionsCreate(cmd *cobra.Command, args []string) error {
	message, err := parseMessage(submissionsMessage)
	if err != nil {
		return err
	}

	// Validate optional email fields
	if submissionsBCCCompleted != "" {
		if err := validation.ValidateEmail(submissionsBCCCompleted); err != nil {
			return fmt.Errorf("invalid bcc-completed email: %w", err)
		}
	}

	if submissionsReplyTo != "" {
		if err := validation.ValidateEmail(submissionsReplyTo); err != nil {
			return fmt.Errorf("invalid reply-to email: %w", err)
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	// Desire path: allow --emails to avoid role syntax entirely.
	if strings.TrimSpace(submissionsEmailsCSV) != "" {
		req := &api.CreateSubmissionsFromEmailsRequest{
			TemplateID: submissionsTemplateID,
			Emails:     submissionsEmailsCSV,
			SendEmail:  submissionsSendEmail,
			Message:    message,
		}

		createdSubmitters, err := client.CreateSubmissionsFromEmails(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create submissions from emails: %w", err)
		}

		outputResult(mode, createdSubmitters, func() {
			if len(createdSubmitters) > 0 {
				fmt.Printf("Created submission %d with %d submitter(s)\n", createdSubmitters[0].SubmissionID, len(createdSubmitters))
			}
		})
		return nil
	}

	if len(submissionsSubmitters) == 0 {
		return fmt.Errorf("either --submitters or --emails is required")
	}

	submitters, err := parseSubmitters(submissionsSubmitters)
	if err != nil {
		return err
	}
	if err := resolveMissingRolesFromTemplate(cmd.Context(), client, submissionsTemplateID, submitters); err != nil {
		return err
	}

	req := &api.CreateSubmissionRequest{
		TemplateID:           submissionsTemplateID,
		Submitters:           submitters,
		SendEmail:            submissionsSendEmail,
		SendSMS:              submissionsSendSMS,
		Message:              message,
		CompletedRedirectURL: submissionsCompletedRedirectURL,
		BCCCompleted:         submissionsBCCCompleted,
		ReplyTo:              submissionsReplyTo,
		ExpireAt:             submissionsExpireAt,
	}

	createdSubmitters, err := client.CreateSubmission(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	outputResult(mode, createdSubmitters, func() {
		if len(createdSubmitters) > 0 {
			fmt.Printf("Created submission %d with %d submitter(s)\n", createdSubmitters[0].SubmissionID, len(createdSubmitters))
			for _, sub := range createdSubmitters {
				displayName := sub.Email
				if sub.Name != "" {
					displayName = sub.Name + " <" + sub.Email + ">"
				}
				fmt.Printf("  - %s (%s): %s\n", displayName, sub.Role, sub.Status)
				if sub.EmbedSrc != "" {
					fmt.Printf("    Sign URL: %s\n", sub.EmbedSrc)
				}
			}
		}
	})

	return nil
}

func runSubmissionsCreatePDF(cmd *cobra.Command, args []string) error {
	submitters, err := parseSubmitters(submissionsSubmitters)
	if err != nil {
		return err
	}
	for i := range submitters {
		if submitters[i].Role == "" {
			submitters[i].Role = "Signer"
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	submission, err := client.CreateSubmissionFromPDF(cmd.Context(), submissionsFile, submitters, submissionsName)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	outputResult(mode, submission, func() {
		fmt.Printf("Created submission %d from PDF\n", submission.ID)
	})

	return nil
}

func runSubmissionsCreateDOCX(cmd *cobra.Command, args []string) error {
	submitters, err := parseSubmitters(submissionsSubmitters)
	if err != nil {
		return err
	}
	for i := range submitters {
		if submitters[i].Role == "" {
			submitters[i].Role = "Signer"
		}
	}

	variables := parseVariables(submissionsVariables)

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	submission, err := client.CreateSubmissionFromDOCX(cmd.Context(), submissionsFile, submitters, submissionsName, variables)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	outputResult(mode, submission, func() {
		fmt.Printf("Created submission %d from DOCX\n", submission.ID)
	})

	return nil
}

func runSubmissionsCreateHTML(cmd *cobra.Command, args []string) error {
	submitters, err := parseSubmitters(submissionsSubmitters)
	if err != nil {
		return err
	}
	for i := range submitters {
		if submitters[i].Role == "" {
			submitters[i].Role = "Signer"
		}
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	submission, err := client.CreateSubmissionFromHTML(cmd.Context(), submissionsHTML, submitters, submissionsName)
	if err != nil {
		return fmt.Errorf("failed to create submission: %w", err)
	}

	outputResult(mode, submission, func() {
		fmt.Printf("Created submission %d from HTML\n", submission.ID)
	})

	return nil
}

func runSubmissionsDocuments(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveSubmissionID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	documents, err := client.GetSubmissionDocuments(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get documents: %w", err)
	}

	outputResult(mode, documents, func() {
		if len(documents) == 0 {
			fmt.Println("No documents found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "NAME\tURL"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, d := range documents {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", d.Name, d.URL); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}
	})

	return nil
}

func runSubmissionsArchive(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveSubmissionID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	if dryRunPreview("archive submission %d", id) {
		return nil
	}

	result, err := client.ArchiveSubmission(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to archive submission: %w", err)
	}

	outputResult(mode, result, func() {
		fmt.Printf("Archived submission %d\n", result.ID)
	})

	return nil
}

func runSubmissionsInit(cmd *cobra.Command, args []string) error {
	submitters, err := parseSubmitters(submissionsSubmitters)
	if err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	if err := resolveMissingRolesFromTemplate(cmd.Context(), client, submissionsTemplateID, submitters); err != nil {
		return err
	}

	req := &api.CreateSubmissionRequest{
		TemplateID: submissionsTemplateID,
		Submitters: submitters,
		SendEmail:  false,
	}

	submission, err := client.InitSubmission(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to initialize submission: %w", err)
	}

	outputResult(mode, submission, func() {
		fmt.Printf("Initialized submission %d\n", submission.ID)
		if len(submission.Submitters) > 0 {
			fmt.Println("Submitters:")
			for _, sub := range submission.Submitters {
				fmt.Printf("  - %s (%s)\n", sub.Email, sub.Role)
				if sub.EmbedSrc != "" {
					fmt.Printf("    Sign URL: %s\n", sub.EmbedSrc)
				}
			}
		}
	})

	return nil
}

func resolveMissingRolesFromTemplate(ctx context.Context, client *api.Client, templateID int, submitters []api.SubmitterRequest) error {
	needsRoles := false
	for _, s := range submitters {
		if s.Role == "" {
			needsRoles = true
			break
		}
	}
	if !needsRoles {
		return nil
	}

	tpl, err := client.GetTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to resolve roles from template %d: %w", templateID, err)
	}

	var roles []string
	for _, r := range tpl.Submitters {
		if r.Name != "" {
			roles = append(roles, r.Name)
		}
	}
	if len(roles) == 0 {
		roles = []string{"Signer"}
	}

	if len(roles) == 1 {
		for i := range submitters {
			if submitters[i].Role == "" {
				submitters[i].Role = roles[0]
			}
		}
		return nil
	}

	// Multiple roles: only auto-assign when it is unambiguous.
	allMissing := true
	for _, s := range submitters {
		if s.Role != "" {
			allMissing = false
			break
		}
	}

	if allMissing {
		if len(submitters) != len(roles) {
			return fmt.Errorf("submitter roles required: template has roles %v, but got %d submitter(s). Use EMAIL:ROLE or provide %d emails", roles, len(submitters), len(roles))
		}
		for i := range submitters {
			submitters[i].Role = roles[i]
		}
		return nil
	}

	// Some roles provided: fill missing by position when possible.
	for i := range submitters {
		if submitters[i].Role != "" {
			continue
		}
		if i >= len(roles) {
			return fmt.Errorf("submitter roles required: template has roles %v. Use EMAIL:ROLE for submitter %d", roles, i+1)
		}
		submitters[i].Role = roles[i]
	}

	return nil
}

func runSubmissionsCreateEmails(cmd *cobra.Command, args []string) error {
	// Validate inputs
	if submissionsEmails == "" {
		return fmt.Errorf("--emails is required and cannot be empty")
	}
	if submissionsTemplateID <= 0 {
		return fmt.Errorf("--template-id must be a positive integer")
	}

	// Validate email list
	if _, err := validation.ValidateEmailList(submissionsEmails); err != nil {
		return fmt.Errorf("invalid email list: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	req := &api.CreateSubmissionsFromEmailsRequest{
		TemplateID: submissionsTemplateID,
		Emails:     submissionsEmails,
		SendEmail:  submissionsSendEmail,
	}

	if submissionsMessageSubject != "" || submissionsMessageBody != "" {
		req.Message = &api.Message{
			Subject: submissionsMessageSubject,
			Body:    submissionsMessageBody,
		}
	}

	submitters, err := client.CreateSubmissionsFromEmails(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create submissions: %w", err)
	}

	outputResult(mode, submitters, func() {
		fmt.Printf("Created %d submitter(s)\n", len(submitters))
		for _, sub := range submitters {
			fmt.Printf("  - %s (%s): %s\n", sub.Email, sub.Role, sub.Status)
			if sub.EmbedSrc != "" {
				fmt.Printf("    Sign URL: %s\n", sub.EmbedSrc)
			}
		}
	})

	return nil
}
