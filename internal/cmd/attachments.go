package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var attachmentsCmd = &cobra.Command{
	Use:     "attachments",
	Aliases: []string{"attachment", "att"},
	Short:   "Manage attachments",
	Long:    `Upload and manage file attachments.`,
}

var attachmentsUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an attachment",
	Long:  `Upload a file as an attachment for use in submissions.`,
	RunE:  runAttachmentsUpload,
}

var attachmentsFile string

func init() {
	rootCmd.AddCommand(attachmentsCmd)
	attachmentsCmd.AddCommand(attachmentsUploadCmd)

	attachmentsUploadCmd.Flags().StringVar(&attachmentsFile, "file", "", "File path to upload (required)")
	mustMarkFlagRequired(attachmentsUploadCmd, "file")
}

func runAttachmentsUpload(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	attachment, err := client.CreateAttachment(cmd.Context(), attachmentsFile)
	if err != nil {
		return fmt.Errorf("failed to upload attachment: %w", err)
	}

	outputResult(mode, attachment, func() {
		fmt.Printf("Uploaded attachment: %s\n", attachment.Name)
		fmt.Printf("ID: %s\n", attachment.ID)
		if attachment.URL != "" {
			fmt.Printf("URL: %s\n", attachment.URL)
		}
	})

	return nil
}
