package cmd

import "testing"

func TestBuildSchemaContainsCommands(t *testing.T) {
	s := buildSchema()
	if s.Name != "docuseal" {
		t.Fatalf("schema name = %q, want %q", s.Name, "docuseal")
	}
	if len(s.Commands) == 0 {
		t.Fatalf("schema commands empty")
	}

	// Expect some well-known command to exist.
	foundTemplates := false
	for _, c := range s.Commands {
		if c.CommandPath == "docuseal templates" || c.CommandPath == "docuseal template" {
			foundTemplates = true
			break
		}
	}
	if !foundTemplates {
		t.Fatalf("expected templates command in schema, got %v commands", len(s.Commands))
	}
}
