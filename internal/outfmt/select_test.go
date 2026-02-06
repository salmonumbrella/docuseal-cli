package outfmt

import "testing"

func TestApplySelect_Object(t *testing.T) {
	in := map[string]any{
		"id":   123,
		"name": "NDA",
		"nested": map[string]any{
			"slug": "nda-1",
		},
	}

	gotAny, err := ApplySelect(in, "id,nested.slug")
	if err != nil {
		t.Fatalf("ApplySelect error: %v", err)
	}
	got := gotAny.(map[string]any)

	if got["id"] != float64(123) { // json roundtrip normalizes numbers
		t.Fatalf("id mismatch: %#v", got["id"])
	}
	nested, ok := got["nested"].(map[string]any)
	if !ok || nested["slug"] != "nda-1" {
		t.Fatalf("nested.slug mismatch: %#v", got["nested"])
	}
	if _, ok := got["name"]; ok {
		t.Fatalf("unexpected key name present: %#v", got)
	}
}

func TestApplySelect_EnvelopePreservesMeta(t *testing.T) {
	in := map[string]any{
		"results": []map[string]any{
			{"id": 1, "name": "A", "x": 9},
			{"id": 2, "name": "B", "x": 8},
		},
		"has_more":   true,
		"next_after": 2,
	}

	gotAny, err := ApplySelect(in, "id,name")
	if err != nil {
		t.Fatalf("ApplySelect error: %v", err)
	}
	got := gotAny.(map[string]any)

	if got["has_more"] != true {
		t.Fatalf("meta lost: %#v", got)
	}
	if got["next_after"] != float64(2) {
		t.Fatalf("meta mismatch: %#v", got["next_after"])
	}

	results, ok := got["results"].([]any)
	if !ok || len(results) != 2 {
		t.Fatalf("results mismatch: %#v", got["results"])
	}
	first := results[0].(map[string]any)
	if _, ok := first["x"]; ok {
		t.Fatalf("unexpected key projected: %#v", first)
	}
}
