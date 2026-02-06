package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{
	Use:     "tools",
	Aliases: []string{"tool"},
	Short:   "PDF utility tools",
	Long:    `Tools for PDF manipulation and verification.`,
}

var toolsMergeCmd = &cobra.Command{
	Use:   "merge-pdfs",
	Short: "Merge multiple PDFs into one",
	Long:  `Merge multiple PDF files into a single PDF document.`,
	RunE:  runToolsMerge,
}

var toolsVerifyCmd = &cobra.Command{
	Use:   "verify-signature",
	Short: "Verify PDF signature",
	Long:  `Verify digital signatures in a PDF document.`,
	RunE:  runToolsVerify,
}

var (
	toolsFiles  string
	toolsOutput string
	toolsFile   string
)

func init() {
	rootCmd.AddCommand(toolsCmd)
	toolsCmd.AddCommand(toolsMergeCmd)
	toolsCmd.AddCommand(toolsVerifyCmd)

	toolsMergeCmd.Flags().StringVar(&toolsFiles, "files", "", "Comma-separated PDF file paths (required)")
	toolsMergeCmd.Flags().StringVar(&toolsOutput, "output", "merged.pdf", "Output file path")
	mustMarkFlagRequired(toolsMergeCmd, "files")

	toolsVerifyCmd.Flags().StringVar(&toolsFile, "file", "", "PDF file to verify (required)")
	mustMarkFlagRequired(toolsVerifyCmd, "file")
}

func runToolsMerge(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	files := strings.Split(toolsFiles, ",")
	for i := range files {
		files[i] = strings.TrimSpace(files[i])
	}

	if len(files) < 2 {
		return fmt.Errorf("at least 2 PDF files required for merge")
	}

	merged, err := client.MergePDFs(cmd.Context(), files)
	if err != nil {
		return fmt.Errorf("failed to merge PDFs: %w", err)
	}

	if err := os.WriteFile(toolsOutput, merged, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	outputResult(mode, map[string]any{
		"files":         files,
		"files_count":   len(files),
		"output_path":   toolsOutput,
		"bytes_written": len(merged),
	}, func() {
		fmt.Printf("Merged %d PDFs into %s\n", len(files), toolsOutput)
	})
	return nil
}

func runToolsVerify(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}
	mode := getOutputMode()

	result, err := client.VerifySignature(cmd.Context(), toolsFile)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	outputResult(mode, result, func() {
		fmt.Printf("Checksum: %s\n", result.ChecksumStatus)
		if len(result.Signatures) == 0 {
			fmt.Println("No signatures found")
			return
		}
		for i, sig := range result.Signatures {
			fmt.Printf("\nSignature %d:\n", i+1)
			fmt.Printf("  Signer: %s\n", sig.SignerName)
			fmt.Printf("  Time: %s\n", sig.SigningTime)
			fmt.Printf("  Type: %s\n", sig.SignatureType)
			if sig.SigningReason != "" {
				fmt.Printf("  Reason: %s\n", sig.SigningReason)
			}
		}
	})

	return nil
}
