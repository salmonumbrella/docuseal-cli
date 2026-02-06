package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:     "templates",
	Aliases: []string{"template", "tpl", "t"},
	Short:   "Manage templates",
	Long:    `List, create, update, and manage DocuSeal templates.`,
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List templates",
	Long:  `List templates with optional filtering by folder or archived status.`,
	Example: `  # List all templates
  docuseal templates list

  # List templates in a specific folder
  docuseal templates list --folder "Contracts"

  # List with JSON output
  docuseal templates list -o json`,
	RunE: runTemplatesList,
}

var templatesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get template details",
	Long:  `Retrieve detailed information about a specific template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesGet,
}

var templatesCreatePDFCmd = &cobra.Command{
	Use:   "create-pdf",
	Short: "Create template from PDF",
	Long:  `Create a new template from a PDF file.`,
	Example: `  # Create template from PDF
  docuseal templates create-pdf --name "Contract" --file contract.pdf

  # Create in a specific folder
  docuseal templates create-pdf --name "NDA" --file nda.pdf --folder "Legal"`,
	RunE: runTemplatesCreatePDF,
}

var templatesCreateDOCXCmd = &cobra.Command{
	Use:   "create-docx",
	Short: "Create template from DOCX",
	Long:  `Create a new template from a DOCX file.`,
	RunE:  runTemplatesCreateDOCX,
}

var templatesCreateHTMLCmd = &cobra.Command{
	Use:   "create-html",
	Short: "Create template from HTML",
	Long:  `Create a new template from HTML content.`,
	RunE:  runTemplatesCreateHTML,
}

var templatesCloneCmd = &cobra.Command{
	Use:   "clone <id>",
	Short: "Clone a template",
	Long:  `Create a copy of an existing template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesClone,
}

var templatesMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge templates",
	Long:  `Merge multiple templates into a single template.`,
	RunE:  runTemplatesMerge,
}

var templatesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update template",
	Long:  `Update a template's name or folder.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesUpdate,
}

var templatesArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive template",
	Long:  `Archive a template (soft delete).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplatesArchive,
}

var templatesUpdateDocumentsCmd = &cobra.Command{
	Use:   "update-documents <id>",
	Short: "Update template documents",
	Long: `Add, replace, or remove documents from an existing template.

When using multiple --file flags, documents are added sequentially and --position/--replace flags are ignored.
The --position and --replace flags only apply when working with a single file or HTML document.`,
	Example: `  # Add a new PDF document to template
  docuseal templates update-documents 123 --file contract.pdf

  # Replace document at position 0
  docuseal templates update-documents 123 --file new.pdf --position 0 --replace

  # Remove document at position 1
  docuseal templates update-documents 123 --position 1 --remove

  # Add HTML document with custom name
  docuseal templates update-documents 123 --html "<p>Agreement</p>" --name "Terms"

  # Merge multiple documents
  docuseal templates update-documents 123 --file doc1.pdf --file doc2.pdf --merge`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplatesUpdateDocuments,
}

// Flags
var (
	templatesLimit       int
	templatesAfter       int
	templatesBefore      int
	templatesFolder      string
	templatesArchived    bool
	templatesName        string
	templatesFile        string
	templatesHTML        string
	templatesIDs         string
	templatesExternalID  string
	templatesSharedLink  string
	templatesHTMLHeader  string
	templatesHTMLFooter  string
	templatesSize        string
	templatesDocFiles    []string
	templatesDocHTML     string
	templatesDocName     string
	templatesDocPosition int
	templatesDocReplace  bool
	templatesDocRemove   bool
	templatesDocMerge    bool
)

func init() {
	rootCmd.AddCommand(templatesCmd)

	templatesCmd.AddCommand(templatesListCmd)
	templatesCmd.AddCommand(templatesGetCmd)
	templatesCmd.AddCommand(templatesCreatePDFCmd)
	templatesCmd.AddCommand(templatesCreateDOCXCmd)
	templatesCmd.AddCommand(templatesCreateHTMLCmd)
	templatesCmd.AddCommand(templatesCloneCmd)
	templatesCmd.AddCommand(templatesMergeCmd)
	templatesCmd.AddCommand(templatesUpdateCmd)
	templatesCmd.AddCommand(templatesArchiveCmd)
	templatesCmd.AddCommand(templatesUpdateDocumentsCmd)

	// List flags
	templatesListCmd.Flags().IntVar(&templatesLimit, "limit", 0, "Maximum number of templates to return")
	templatesListCmd.Flags().IntVar(&templatesAfter, "after", 0, "Pagination cursor, get IDs greater than value")
	templatesListCmd.Flags().IntVar(&templatesBefore, "before", 0, "Pagination cursor, get IDs less than value")
	templatesListCmd.Flags().StringVar(&templatesFolder, "folder", "", "Filter by folder name")
	templatesListCmd.Flags().BoolVar(&templatesArchived, "archived", false, "Include archived templates")

	// Create PDF flags
	templatesCreatePDFCmd.Flags().StringVar(&templatesName, "name", "", "Template name (required)")
	templatesCreatePDFCmd.Flags().StringVar(&templatesFile, "file", "", "PDF file path (required)")
	templatesCreatePDFCmd.Flags().StringVar(&templatesFolder, "folder", "", "Folder name")
	templatesCreatePDFCmd.Flags().StringVar(&templatesExternalID, "external-id", "", "App-specific identifier")
	templatesCreatePDFCmd.Flags().StringVar(&templatesSharedLink, "shared-link", "", "Enable/disable shared link (true/false, default: true)")
	mustMarkFlagRequired(templatesCreatePDFCmd, "name")
	mustMarkFlagRequired(templatesCreatePDFCmd, "file")

	// Create DOCX flags
	templatesCreateDOCXCmd.Flags().StringVar(&templatesName, "name", "", "Template name (required)")
	templatesCreateDOCXCmd.Flags().StringVar(&templatesFile, "file", "", "DOCX file path (required)")
	templatesCreateDOCXCmd.Flags().StringVar(&templatesFolder, "folder", "", "Folder name")
	templatesCreateDOCXCmd.Flags().StringVar(&templatesExternalID, "external-id", "", "App-specific identifier")
	templatesCreateDOCXCmd.Flags().StringVar(&templatesSharedLink, "shared-link", "", "Enable/disable shared link (true/false, default: true)")
	mustMarkFlagRequired(templatesCreateDOCXCmd, "name")
	mustMarkFlagRequired(templatesCreateDOCXCmd, "file")

	// Create HTML flags
	templatesCreateHTMLCmd.Flags().StringVar(&templatesName, "name", "", "Template name (required)")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesHTML, "html", "", "HTML content (required)")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesFolder, "folder", "", "Folder name")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesExternalID, "external-id", "", "App-specific identifier")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesSharedLink, "shared-link", "", "Enable/disable shared link (true/false, default: true)")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesHTMLHeader, "html-header", "", "Header HTML for every page")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesHTMLFooter, "html-footer", "", "Footer HTML for every page")
	templatesCreateHTMLCmd.Flags().StringVar(&templatesSize, "size", "", "Page size (Letter, Legal, A4, A5, etc.)")
	mustMarkFlagRequired(templatesCreateHTMLCmd, "name")
	mustMarkFlagRequired(templatesCreateHTMLCmd, "html")

	// Clone flags
	templatesCloneCmd.Flags().StringVar(&templatesName, "name", "", "New template name")
	templatesCloneCmd.Flags().StringVar(&templatesFolder, "folder", "", "Folder name")

	// Merge flags
	templatesMergeCmd.Flags().StringVar(&templatesIDs, "ids", "", "Comma-separated template IDs (required)")
	templatesMergeCmd.Flags().StringVar(&templatesName, "name", "", "Merged template name (required)")
	templatesMergeCmd.Flags().StringVar(&templatesFolder, "folder", "", "Folder name")
	mustMarkFlagRequired(templatesMergeCmd, "ids")
	mustMarkFlagRequired(templatesMergeCmd, "name")

	// Update flags
	templatesUpdateCmd.Flags().StringVar(&templatesName, "name", "", "New template name")
	templatesUpdateCmd.Flags().StringVar(&templatesFolder, "folder", "", "New folder name")

	// Update documents flags
	templatesUpdateDocumentsCmd.Flags().StringArrayVar(&templatesDocFiles, "file", []string{}, "Document file path (PDF/DOCX, can be specified multiple times)")
	templatesUpdateDocumentsCmd.Flags().StringVar(&templatesDocHTML, "html", "", "HTML content for document")
	templatesUpdateDocumentsCmd.Flags().StringVar(&templatesDocName, "name", "", "Document name")
	templatesUpdateDocumentsCmd.Flags().IntVar(&templatesDocPosition, "position", 0, "Document position (0-based index)")
	templatesUpdateDocumentsCmd.Flags().BoolVar(&templatesDocReplace, "replace", false, "Replace document at position")
	templatesUpdateDocumentsCmd.Flags().BoolVar(&templatesDocRemove, "remove", false, "Remove document at position")
	templatesUpdateDocumentsCmd.Flags().BoolVar(&templatesDocMerge, "merge", false, "Merge all documents")
}

func runTemplatesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	limit := templatesLimit
	reqLimit := limit
	if ((mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta)) && limit > 0 {
		reqLimit = limit + 1
	}

	templates, err := client.ListTemplates(cmd.Context(), reqLimit, templatesFolder, templatesArchived, templatesAfter, templatesBefore)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	// Machine-friendly envelope / meta.
	if (mode == outfmt.JSON && !bareJSON) || (mode == outfmt.NDJSON && withMeta) {
		out := templates
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
			env := makeListEnvelope(out, len(out), limit, templatesAfter, templatesBefore, hasMore, nextAfter, nextBefore)
			outputResult(mode, env, func() {})
			return nil
		}

		// NDJSON with trailing meta line.
		meta := map[string]any{
			"_meta": map[string]any{
				"count":       len(out),
				"limit":       limit,
				"after":       templatesAfter,
				"before":      templatesBefore,
				"has_more":    hasMore,
				"next_after":  nextAfter,
				"next_before": nextBefore,
			},
		}
		stream := make([]any, 0, len(out)+1)
		for _, t := range out {
			stream = append(stream, t)
		}
		stream = append(stream, meta)
		outputResult(mode, stream, func() {})
		return nil
	}

	outputResult(mode, templates, func() {
		if len(templates) == 0 {
			fmt.Println("No templates found")
			return
		}
		w := newTabWriter()
		if _, err := fmt.Fprintln(w, "ID\tNAME\tFOLDER\tCREATED"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
		for _, t := range templates {
			if _, err := fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				t.ID,
				truncateString(t.Name, 40),
				t.FolderName,
				formatTime(t.CreatedAt),
			); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			}
		}
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		}

		// Pagination hint
		if templatesLimit > 0 && len(templates) == templatesLimit {
			fmt.Fprintf(os.Stderr, "\n# More results may be available. Use --after %d to see next page.\n", templates[len(templates)-1].ID)
		}
	})

	return nil
}

func runTemplatesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveTemplateID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	template, err := client.GetTemplate(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("ID: %d\n", template.ID)
		fmt.Printf("Name: %s\n", template.Name)
		fmt.Printf("Slug: %s\n", template.Slug)
		fmt.Printf("Folder: %s\n", template.FolderName)
		fmt.Printf("Created: %s\n", formatTime(template.CreatedAt))
		fmt.Printf("Updated: %s\n", formatTime(template.UpdatedAt))
		if template.ArchivedAt != nil {
			fmt.Printf("Archived: %s\n", formatTimePtr(template.ArchivedAt))
		}
		if len(template.Fields) > 0 {
			fmt.Printf("Fields: %d\n", len(template.Fields))
		}
		if len(template.Submitters) > 0 {
			fmt.Printf("Roles: ")
			for i, s := range template.Submitters {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s", s.Name)
			}
			fmt.Println()
		}
	})

	return nil
}

func runTemplatesCreatePDF(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	var sharedLink *bool
	if templatesSharedLink != "" {
		val, err := strconv.ParseBool(templatesSharedLink)
		if err != nil {
			return fmt.Errorf("invalid shared-link value: must be true or false")
		}
		sharedLink = &val
	}

	template, err := client.CreateTemplateFromPDF(cmd.Context(), templatesName, templatesFile, templatesFolder, templatesExternalID, sharedLink)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Created template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesCreateDOCX(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	var sharedLink *bool
	if templatesSharedLink != "" {
		val, err := strconv.ParseBool(templatesSharedLink)
		if err != nil {
			return fmt.Errorf("invalid shared-link value: must be true or false")
		}
		sharedLink = &val
	}

	template, err := client.CreateTemplateFromDOCX(cmd.Context(), templatesName, templatesFile, templatesFolder, templatesExternalID, sharedLink)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Created template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesCreateHTML(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	var sharedLink *bool
	if templatesSharedLink != "" {
		val, err := strconv.ParseBool(templatesSharedLink)
		if err != nil {
			return fmt.Errorf("invalid shared-link value: must be true or false")
		}
		sharedLink = &val
	}

	template, err := client.CreateTemplateFromHTML(cmd.Context(), templatesName, templatesHTML, templatesFolder, templatesExternalID, templatesHTMLHeader, templatesHTMLFooter, templatesSize, sharedLink)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Created template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesClone(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveTemplateID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	template, err := client.CloneTemplate(cmd.Context(), id, templatesName, templatesFolder)
	if err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Cloned template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesMerge(cmd *cobra.Command, args []string) error {
	ids, err := parseIntSlice(templatesIDs)
	if err != nil {
		return fmt.Errorf("invalid template IDs: %w", err)
	}

	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	template, err := client.MergeTemplates(cmd.Context(), ids, templatesName, templatesFolder)
	if err != nil {
		return fmt.Errorf("failed to merge templates: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Merged into template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesUpdate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveTemplateID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	template, err := client.UpdateTemplate(cmd.Context(), id, templatesName, templatesFolder)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Updated template %d: %s\n", template.ID, template.Name)
	})

	return nil
}

func runTemplatesArchive(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveTemplateID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	if dryRunPreview("archive template %d", id) {
		return nil
	}

	result, err := client.ArchiveTemplate(cmd.Context(), id)
	if err != nil {
		return fmt.Errorf("failed to archive template: %w", err)
	}

	outputResult(mode, result, func() {
		fmt.Printf("Archived template %d\n", result.ID)
	})

	return nil
}

func runTemplatesUpdateDocuments(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	id, err := resolveTemplateID(cmd.Context(), client, args[0])
	if err != nil {
		return err
	}

	// Validate position
	if templatesDocPosition < 0 {
		return fmt.Errorf("position must be non-negative")
	}

	// Build document operations
	var operations []api.TemplateDocumentOperation

	// Handle file uploads
	for i, filePath := range templatesDocFiles {
		op, err := buildDocumentOperationFromFile(filePath, i)
		if err != nil {
			return err
		}
		if templatesDocName != "" && len(templatesDocFiles) == 1 {
			op.Name = templatesDocName
		}
		operations = append(operations, op)
	}

	// Handle HTML content
	if templatesDocHTML != "" {
		if err := api.ValidateHTMLContent(templatesDocHTML); err != nil {
			return err
		}
		op := api.TemplateDocumentOperation{
			HTML:     templatesDocHTML,
			Name:     templatesDocName,
			Position: templatesDocPosition,
			Replace:  templatesDocReplace,
			Remove:   templatesDocRemove,
		}
		operations = append(operations, op)
	}

	// Handle remove operation (no file or HTML)
	if templatesDocRemove && len(operations) == 0 {
		if dryRunPreview("remove document at position %d from template %d", templatesDocPosition, id) {
			return nil
		}
		op := api.TemplateDocumentOperation{
			Position: templatesDocPosition,
			Remove:   true,
		}
		operations = append(operations, op)
	}

	// Validate we have at least one operation
	if len(operations) == 0 {
		return fmt.Errorf("no document operations specified (use --file, --html, or --remove)")
	}

	// Apply position/replace flags to file operations if only one file
	if len(templatesDocFiles) == 1 && templatesDocHTML == "" {
		operations[0].Position = templatesDocPosition
		operations[0].Replace = templatesDocReplace
	}

	req := &api.UpdateTemplateDocumentsRequest{
		Documents: operations,
		Merge:     templatesDocMerge,
	}

	template, err := client.UpdateTemplateDocuments(cmd.Context(), id, req)
	if err != nil {
		return fmt.Errorf("failed to update template documents: %w", err)
	}

	outputResult(mode, template, func() {
		fmt.Printf("Updated documents for template %d: %s\n", template.ID, template.Name)
		if template.DocumentsCount > 0 {
			fmt.Printf("Total documents: %d\n", template.DocumentsCount)
		}
	})

	return nil
}

func buildDocumentOperationFromFile(filePath string, index int) (api.TemplateDocumentOperation, error) {
	// Validate file size
	if err := api.ValidateFileSize(filePath); err != nil {
		return api.TemplateDocumentOperation{}, err
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return api.TemplateDocumentOperation{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(fileData)

	// Determine MIME type
	mimeType := "application/pdf"
	if strings.HasSuffix(strings.ToLower(filePath), ".docx") {
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	op := api.TemplateDocumentOperation{
		File:     fmt.Sprintf("data:%s;base64,%s", mimeType, encoded),
		Position: index,
	}

	return op, nil
}
