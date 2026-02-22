package outfmt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ApplySelect projects v to a smaller JSON shape based on selectStr.
// selectStr is a comma-separated list of JSON keys or dot paths.
//
// Behavior:
//   - If v is a list-envelope object (has "results"), selection is applied to each result element,
//     and envelope metadata is preserved.
//   - If v is an array/slice, selection is applied per element (when elements are objects).
//   - Missing paths are ignored (no error).
func ApplySelect(v any, selectStr string) (any, error) {
	fields := parseSelect(selectStr)
	if len(fields) == 0 {
		return v, nil
	}

	anyVal, err := normalizeToAny(v)
	if err != nil {
		return nil, err
	}

	switch tv := anyVal.(type) {
	case map[string]any:
		// Keep meta objects intact in NDJSON streams.
		if _, ok := tv["_meta"]; ok {
			return tv, nil
		}

		// Special case: list envelope.
		if results, ok := tv["results"]; ok {
			if arr, ok := results.([]any); ok {
				outArr := make([]any, 0, len(arr))
				for _, el := range arr {
					if m, ok := el.(map[string]any); ok {
						if _, ok := m["_meta"]; ok {
							outArr = append(outArr, m)
							continue
						}
						outArr = append(outArr, projectObject(m, fields))
					} else {
						outArr = append(outArr, el)
					}
				}
				tv["results"] = outArr
			}
			return tv, nil
		}
		return projectObject(tv, fields), nil
	case []any:
		out := make([]any, 0, len(tv))
		for _, el := range tv {
			if m, ok := el.(map[string]any); ok {
				if _, ok := m["_meta"]; ok {
					out = append(out, m)
					continue
				}
				out = append(out, projectObject(m, fields))
			} else {
				out = append(out, el)
			}
		}
		return out, nil
	default:
		// Primitive or unknown: nothing to project.
		return anyVal, nil
	}
}

func parseSelect(s string) [][]string {
	raw := strings.Split(s, ",")
	var out [][]string
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		parts := strings.Split(r, ".")
		var cleaned []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cleaned = append(cleaned, p)
			}
		}
		if len(cleaned) > 0 {
			out = append(out, cleaned)
		}
	}
	return out
}

func normalizeToAny(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal for selection: %w", err)
	}
	var out any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal for selection: %w", err)
	}
	return out, nil
}

func projectObject(in map[string]any, fields [][]string) map[string]any {
	out := map[string]any{}
	for _, path := range fields {
		if val, ok := getPath(in, path); ok {
			setPath(out, path, val)
		}
	}
	return out
}

func getPath(in map[string]any, path []string) (any, bool) {
	var cur any = in
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func setPath(out map[string]any, path []string, val any) {
	if len(path) == 0 {
		return
	}
	cur := out
	for i := 0; i < len(path)-1; i++ {
		k := path[i]
		next, ok := cur[k]
		if ok {
			if nm, ok := next.(map[string]any); ok {
				cur = nm
				continue
			}
		}
		nm := map[string]any{}
		cur[k] = nm
		cur = nm
	}
	cur[path[len(path)-1]] = val
}
