# 📄 DocuSeal CLI — Document signing in your terminal.

DocuSeal in your terminal. Manage templates, submissions, submitters, webhooks, and document signing workflows.

## Features

- **Attachments** - upload files for document workflows
- **Authentication** - authenticate once, credentials stored securely in OS keychain
- **Events** - view form and submission event history
- **Multiple instances** - manage multiple DocuSeal instances (cloud or self-hosted)
- **PDF tools** - merge PDFs and verify signatures locally
- **Submissions** - create signing workflows from templates or one-off documents
- **Submitters** - manage signers, track status, programmatically complete signing
- **Templates** - create, clone, merge, and update document templates from PDF/DOCX/HTML
- **Webhooks** - configure event notifications for automation

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/docuseal-cli
```

### From Source

```bash
git clone https://github.com/salmonumbrella/docuseal-cli.git
cd docuseal-cli
make build
```

The binary will be available at `./bin/docuseal`.

## Quick Start

### 1. Authenticate

```bash
docuseal auth login
# Opens a browser for authentication and stores credentials in your OS keychain

# Or authenticate non-interactively:
docuseal auth login --url https://docuseal.example.com --api-key YOUR_KEY
```

### 2. Test Authentication

```bash
docuseal auth status
```

### 3. List Templates

```bash
docuseal templates list
```

### 4. Create a Submission

```bash
docuseal submissions create \
  --template-id 123 \
  --submitters "john@example.com:Signer" \
  --send-email
```

## Configuration

### Environment Variables

- `DOCUSEAL_OUTPUT` - Output format: `text` (default), `json`, or `ndjson`
- `DOCUSEAL_COLOR` - Color mode: `auto` (default), `always`, or `never`
- `DOCUSEAL_TIMEOUT` - HTTP request timeout (e.g. `30s`, `2m`)
- `DOCUSEAL_RETRIES` - Max retries for rate-limited requests (HTTP 429)
- `DOCUSEAL_RETRY_BASE_DELAY` - Base delay for 429 backoff (e.g. `1s`)
- `DOCUSEAL_INSECURE_SKIP_VERIFY` - Set to `true` to skip TLS verification (self-signed certs)
- `NO_COLOR` - Set to any value to disable colors (standard convention)

## Security

### Credential Storage

Credentials are stored securely in your system's keychain:
- **macOS**: Keychain Access
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

## Commands

### Authentication

```bash
docuseal auth login                      # Store credentials (browser by default)
docuseal auth login --url ... --api-key ...  # Non-interactive login
docuseal auth status                     # Show current configuration and connectivity
docuseal auth whoami                     # Show authenticated user info
docuseal auth logout                     # Remove stored credentials

# Desire paths (aliases):
docuseal login
docuseal status
docuseal whoami
docuseal logout
```

### Templates

```bash
docuseal templates list [--limit <n>] [--folder <name>]
docuseal templates get <templateId>
docuseal templates create-pdf --name <name> --file <path.pdf>
docuseal templates create-docx --name <name> --file <path.docx>
docuseal templates create-html --name <name> --html <html>
docuseal templates clone <templateId> --name <name>
docuseal templates merge --ids <id1,id2> --name <name>
docuseal templates update <templateId> [--name <name>] [--folder <folder>]
docuseal templates update-documents <templateId> --file <path>
docuseal templates archive <templateId>
```

### Submissions

```bash
docuseal submissions list [--template-id <id>] [--status <status>]
docuseal submissions get <submissionId>
docuseal submissions create --template-id <id> --submitters <email:role> [--send-email]
docuseal submissions create-pdf --file <path.pdf> --submitters <email:role>
docuseal submissions create-docx --file <path.docx> --submitters <email:role>
docuseal submissions create-html --html <html> --submitters <email:role>
docuseal submissions create-emails --template-id <id> --emails <email1,email2>
docuseal submissions init --template-id <id> --submitters <email:role>  # Don't send emails
docuseal submissions documents <submissionId>                           # Get signed documents
docuseal submissions archive <submissionId>
```

### Submitters

```bash
docuseal submitters list [--submission-id <id>]
docuseal submitters get <submitterId>                  # Includes signing URL
docuseal submitters update <submitterId> [--email <email>] [--name <name>]
docuseal submitters update <submitterId> --completed   # Programmatically sign
docuseal submitters update <submitterId> --send-email  # Send notification
```

### Webhooks

```bash
docuseal webhooks list
docuseal webhooks get <webhookId>
docuseal webhooks create --url <url> --events <event1,event2>
docuseal webhooks update <webhookId> [--url <url>] [--events <events>]
docuseal webhooks delete <webhookId>
```

### Attachments

```bash
docuseal attachments upload --file <path>
```

### Events

```bash
docuseal events list [--limit <n>]
```

### PDF Tools

```bash
docuseal tools merge-pdfs --files <file1.pdf,file2.pdf> --output <merged.pdf>
docuseal tools verify-signature --file <signed.pdf>
```

## Output Formats

### Text

Human-readable tables with colors and formatting:

```bash
$ docuseal templates list
ID      NAME              FOLDER      CREATED
123     Employment        Contracts   2024-01-15
456     NDA               Legal       2024-01-20

$ docuseal submissions list
ID      TEMPLATE    STATUS     SUBMITTERS    CREATED
789     123         pending    1/2           2024-01-25
```

### JSON

Machine-readable output:

```bash
$ docuseal templates list --output json
{
  "results": [
    {
      "id": 123,
      "name": "Employment Contract",
      "folder_name": "Contracts",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "count": 1,
  "limit": 0,
  "after": 0,
  "before": 0,
  "has_more": false
}
```

Data goes to stdout, errors and progress to stderr for clean piping.

Use `--bare` to get the old list shape (arrays) if you want:

```bash
docuseal templates list --output json --bare
```

### NDJSON (jsonl)

Stream-friendly output (one JSON object per line):

```bash
$ docuseal templates list --output ndjson
{"id":123,"slug":"...","name":"..."}
{"id":456,"slug":"...","name":"..."}
```

### Compact JSON

Reduce token/byte usage:

```bash
$ docuseal templates list --output json --compact-json
{"results":[{"id":123,"name":"..."},{"id":456,"name":"..."}],"count":2,"limit":0,"after":0,"before":0,"has_more":false}
```

### Field Projection

Reduce output size for agents/scripts:

```bash
docuseal templates list -o json --select id,name
docuseal templates list -o ndjson --select id,name
```

### CLI Schema (Tool Routers)

Let an agent or tool router discover the command surface:

```bash
docuseal schema -o json
docuseal help submissions create --json -o json
```

## Examples

### Complete Signing Workflow

```bash
# Create a submission
SUBMITTERS=$(docuseal submissions create \
  --template-id 123 \
  --submitters "client@example.com:Client" \
  --output json)

# Extract IDs
SUBMISSION_ID=$(echo "$SUBMITTERS" | jq -r '.[0].submission_id')
SUBMITTER_ID=$(echo "$SUBMITTERS" | jq -r '.[0].id')

# Check submitter status
docuseal submitters get "$SUBMITTER_ID"

# Get signed documents after completion
docuseal submissions documents "$SUBMISSION_ID" --output json | jq -r '.[].url'
```

### Auto-Sign Workflow (API Signing)

```bash
# Create submission without sending emails
SUBMISSION=$(docuseal submissions init \
  --template-id 123 \
  --submitters "system@example.com:System" \
  --output json)

# Programmatically complete signing
SUBMITTER_ID=$(echo "$SUBMISSION" | jq -r '.submitters[0].id')
docuseal submitters update "$SUBMITTER_ID" --completed
```

### Template Management

```bash
# Clone existing template
docuseal templates clone 123 --name "Contract Copy - 2024"

# Merge multiple templates
docuseal templates merge --ids 123,456 --name "Combined Agreement"

# Update template documents
docuseal templates update-documents 123 --file ./updated-contract.pdf
```

### Automation

```bash
# Get all template IDs
docuseal templates list --output json | jq -r '.results[].id'

# Filter pending submissions
docuseal submissions list --status pending --output json | jq -r '.count'

# Extract signing URLs
docuseal submitters list --submission-id 789 --output json | jq -r '.results[].embed_src'
```

### Dry-Run Mode

Preview mutations before executing:

```bash
docuseal templates archive 123 --dry-run
# Output: [DRY RUN] Would archive template 123

docuseal submissions create --dry-run \
  --template-id 123 \
  --submitters "test@example.com:Signer"
# Shows what would be created without making API call
```

## Global Flags

All commands support these flags:

- `--output <format>` - Output format: `text`, `json`, or `ndjson` (default: text)
- `--compact-json` - Compact JSON encoding (smaller output for agents/scripts)
- `--select <fields>` - Project JSON output to specific fields (comma-separated)
- `--bare` - For list commands: output arrays in JSON instead of an envelope
- `--meta` - For NDJSON list output: append a final `{"_meta": ...}` line
- `--timeout <duration>` - HTTP request timeout
- `--retries <n>` - Max retries for rate-limited requests (HTTP 429)
- `--retry-base-delay <duration>` - Base delay for 429 backoff
- `--insecure-skip-verify` - Skip TLS certificate verification (insecure)
- `--color <mode>` - Color mode: `auto`, `always`, or `never` (default: auto)
- `--dry-run` - Preview destructive operations without executing
- `--quiet` - Suppress non-essential warnings and progress output
- `--help` - Show help for any command
- `--version` - Show version information

## Exit Codes

- `1` unknown error
- `2` validation error
- `3` authentication error
- `4` rate limited
- `5` not configured
- `6` circuit breaker open
- `7` timeout

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
# macOS (Homebrew):
docuseal completion bash > $(brew --prefix)/etc/bash_completion.d/docuseal

# Linux:
docuseal completion bash > /etc/bash_completion.d/docuseal

# Or source directly:
source <(docuseal completion bash)
```

### Zsh

```zsh
docuseal completion zsh > "${fpath[1]}/_docuseal"
```

### Fish

```fish
docuseal completion fish > ~/.config/fish/completions/docuseal.fish
```

### PowerShell

```powershell
docuseal completion powershell | Out-String | Invoke-Expression
```

## Development

After cloning, install git hooks:

```bash
make setup
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

## License

MIT

## Links

- [DocuSeal API Documentation](https://www.docuseal.co/docs/api)
