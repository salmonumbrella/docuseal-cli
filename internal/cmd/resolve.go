package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
)

func resolveTemplateID(ctx context.Context, client *api.Client, ident string) (int, error) {
	// Fast path: numeric ID or URL containing ID.
	if id, err := parseIDArg(ident); err == nil {
		return id, nil
	}

	needle := strings.ToLower(strings.TrimSpace(ident))
	if needle == "" {
		return 0, fmt.Errorf("empty template identifier")
	}

	scan := func(includeArchived bool) (int, error) {
		const pageLimit = 100
		after := 0

		var slugMatches []api.Template
		var nameMatches []api.Template
		var fuzzyMatches []api.Template

		for pages := 0; pages < 200; pages++ {
			items, err := client.ListTemplates(ctx, pageLimit, "", includeArchived, after, 0)
			if err != nil {
				return 0, fmt.Errorf("failed to resolve template %q: %w", ident, err)
			}
			if len(items) == 0 {
				break
			}

			for _, t := range items {
				if strings.EqualFold(t.Slug, ident) {
					slugMatches = append(slugMatches, t)
					continue
				}
				if strings.EqualFold(t.Name, ident) {
					nameMatches = append(nameMatches, t)
					continue
				}
				if needle != "" {
					if strings.Contains(strings.ToLower(t.Slug), needle) || strings.Contains(strings.ToLower(t.Name), needle) {
						fuzzyMatches = append(fuzzyMatches, t)
					}
				}
			}

			last := items[len(items)-1].ID
			if last <= after {
				break
			}
			after = last
		}

		if id, err := pickResolvedTemplate(slugMatches, ident, "slug"); err != nil {
			return 0, err
		} else if id != 0 {
			return id, nil
		}
		if id, err := pickResolvedTemplate(nameMatches, ident, "name"); err != nil {
			return 0, err
		} else if id != 0 {
			return id, nil
		}
		if id, err := pickResolvedTemplate(fuzzyMatches, ident, "partial match"); err != nil {
			return 0, err
		} else if id != 0 {
			return id, nil
		}

		return 0, nil
	}

	// Try without archived first to avoid depending on API semantics.
	if id, err := scan(false); err != nil {
		return 0, err
	} else if id != 0 {
		return id, nil
	}
	// Then include archived.
	if id, err := scan(true); err != nil {
		return 0, err
	} else if id != 0 {
		return id, nil
	}

	return 0, fmt.Errorf("template %q not found (try numeric ID, URL, exact slug, or exact name)", ident)
}

func pickResolvedTemplate(matches []api.Template, ident string, kind string) (int, error) {
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) > 1 {
		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "template identifier %q is ambiguous (%s matched multiple templates):\n", ident, kind)
		for i, m := range matches {
			if i >= 10 {
				b.WriteString("  ...\n")
				break
			}
			_, _ = fmt.Fprintf(&b, "  - id=%d slug=%q name=%q\n", m.ID, m.Slug, m.Name)
		}
		return 0, fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
	}
	return 0, nil
}

func resolveSubmissionID(ctx context.Context, client *api.Client, ident string) (int, error) {
	if id, err := parseIDArg(ident); err == nil {
		return id, nil
	}

	slug := strings.TrimSpace(ident)
	if slug == "" {
		return 0, fmt.Errorf("empty submission identifier")
	}

	// First try active submissions.
	items, err := client.ListSubmissions(ctx, 2, 0, "", "", slug, "", false, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve submission %q: %w", ident, err)
	}
	if len(items) == 0 {
		// Then try archived.
		items, err = client.ListSubmissions(ctx, 2, 0, "", "", slug, "", true, 0, 0)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve submission %q: %w", ident, err)
		}
	}

	if len(items) == 1 {
		return items[0].ID, nil
	}
	if len(items) > 1 {
		return 0, fmt.Errorf("submission slug %q matched multiple submissions; use numeric ID", ident)
	}
	return 0, fmt.Errorf("submission %q not found (try numeric ID, URL, or exact slug)", ident)
}
