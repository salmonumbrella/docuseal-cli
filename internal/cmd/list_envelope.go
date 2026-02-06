package cmd

// Helpers to produce machine-friendly list envelopes for JSON output.

func makeListEnvelope(results any, count int, limit, after, before int, hasMore bool, nextAfter, nextBefore int) map[string]any {
	env := map[string]any{
		"results":  results,
		"count":    count,
		"limit":    limit,
		"after":    after,
		"before":   before,
		"has_more": hasMore,
	}
	if nextAfter != 0 {
		env["next_after"] = nextAfter
	}
	if nextBefore != 0 {
		env["next_before"] = nextBefore
	}
	return env
}
