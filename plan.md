# Plan: Agent-Friendly CLI Improvements (DocuSeal CLI)

## Goals
- Reduce navigation and guessing for agents (support name/slug resolution).
- Reduce token/byte output by default for agents (server results + client projection).
- Make pagination machine-visible (metadata in JSON, streaming-friendly NDJSON option).

## Work Items
1. Identifier Resolution
   - Templates: accept `<id|url|slug|name>` for commands that take a template identifier.
   - Submissions: accept `<id|url|slug>` for commands that take a submission identifier.
   - Keep existing numeric/URL behavior; add fallback resolution only when non-numeric.

2. Output Projection
   - Add global `--select` (comma-separated JSON keys / dot paths) applied to JSON/NDJSON outputs.
   - Apply selection to list envelopes by projecting `results` only, preserving metadata.

3. Pagination Metadata
   - For list commands, default JSON output becomes an envelope:
     - `{ "results": [...], "count": n, "limit": ..., "after": ..., "before": ..., "has_more": ..., "next_after": ..., "next_before": ... }`
   - Add global `--bare` to keep the old JSON shape (arrays) for backward compatibility.
   - For NDJSON, keep resource-per-line; optionally append a final meta line when `--meta` is set.

## Validation
- `go test ./...`
- Spot-check CLI help and output behavior for:
  - `--output json|ndjson`, `--select`, `--bare`, `--meta`
  - Template name/slug resolution

