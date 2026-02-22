package outfmt

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMode  Mode
		wantError bool
	}{
		{
			name:      "text mode",
			input:     "text",
			wantMode:  Text,
			wantError: false,
		},
		{
			name:      "json mode",
			input:     "json",
			wantMode:  JSON,
			wantError: false,
		},
		{
			name:      "empty string defaults to text",
			input:     "",
			wantMode:  Text,
			wantError: false,
		},
		{
			name:      "invalid mode",
			input:     "xml",
			wantMode:  Text,
			wantError: true,
		},
		{
			name:      "invalid mode uppercase",
			input:     "JSON",
			wantMode:  Text,
			wantError: true,
		},
		{
			name:      "invalid mode with spaces",
			input:     " json ",
			wantMode:  Text,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := Parse(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Parse(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
				return
			}
			if mode != tt.wantMode {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, mode, tt.wantMode)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid output format") {
				t.Errorf("Parse(%q) error message = %q, want error containing 'invalid output format'", tt.input, err.Error())
			}
		})
	}
}

func TestWithMode(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{
			name: "text mode",
			mode: Text,
		},
		{
			name: "json mode",
			mode: JSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithMode(ctx, tt.mode)

			got := ModeFromContext(ctx)
			if got != tt.mode {
				t.Errorf("WithMode/ModeFromContext() = %v, want %v", got, tt.mode)
			}
		})
	}
}

func TestModeFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		want     Mode
	}{
		{
			name: "context with text mode",
			setupCtx: func() context.Context {
				return WithMode(context.Background(), Text)
			},
			want: Text,
		},
		{
			name: "context with json mode",
			setupCtx: func() context.Context {
				return WithMode(context.Background(), JSON)
			},
			want: JSON,
		},
		{
			name: "context without mode defaults to text",
			setupCtx: func() context.Context {
				return context.Background()
			},
			want: Text,
		},
		{
			name: "context with wrong type defaults to text",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), contextKey{}, "invalid")
			},
			want: Text,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			got := ModeFromContext(ctx)
			if got != tt.want {
				t.Errorf("ModeFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
		want bool
	}{
		{
			name: "json mode returns true",
			mode: JSON,
			want: true,
		},
		{
			name: "text mode returns false",
			mode: Text,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithMode(context.Background(), tt.mode)
			got := IsJSON(ctx)
			if got != tt.want {
				t.Errorf("IsJSON() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("context without mode returns false", func(t *testing.T) {
		ctx := context.Background()
		got := IsJSON(ctx)
		if got != false {
			t.Errorf("IsJSON() on empty context = %v, want false", got)
		}
	})
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 42,
			},
			want: `{
  "name": "test",
  "value": 42
}
`,
			wantErr: false,
		},
		{
			name: "map",
			input: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			want: `{
  "key1": "value1",
  "key2": 123,
  "key3": true
}
`,
			wantErr: false,
		},
		{
			name: "slice",
			input: []string{
				"item1",
				"item2",
				"item3",
			},
			want: `[
  "item1",
  "item2",
  "item3"
]
`,
			wantErr: false,
		},
		{
			name:  "nil value",
			input: nil,
			want: `null
`,
			wantErr: false,
		},
		{
			name:  "empty struct",
			input: struct{}{},
			want: `{}
`,
			wantErr: false,
		},
		{
			name:  "empty slice",
			input: []string{},
			want: `[]
`,
			wantErr: false,
		},
		{
			name:  "empty map",
			input: map[string]any{},
			want: `{}
`,
			wantErr: false,
		},
		{
			name: "nested structure",
			input: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
					"tags": []string{"admin", "user"},
				},
			},
			want: `{
  "user": {
    "age": 30,
    "name": "John",
    "tags": [
      "admin",
      "user"
    ]
  }
}
`,
			wantErr: false,
		},
		{
			name:  "string value",
			input: "simple string",
			want: `"simple string"
`,
			wantErr: false,
		},
		{
			name:  "number value",
			input: 42,
			want: `42
`,
			wantErr: false,
		},
		{
			name:  "boolean value",
			input: true,
			want: `true
`,
			wantErr: false,
		},
		{
			name: "struct with json tags",
			input: struct {
				PublicField  string `json:"public"`
				privateField string
				OmitEmpty    string `json:"omit,omitempty"`
				Renamed      string `json:"new_name"`
			}{
				PublicField:  "visible",
				privateField: "hidden",
				OmitEmpty:    "",
				Renamed:      "renamed_value",
			},
			want: `{
  "public": "visible",
  "new_name": "renamed_value"
}
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := WriteJSON(buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("WriteJSON() output mismatch:\ngot:\n%s\nwant:\n%s", got, tt.want)
				}

				// Verify output is valid JSON
				var check any
				if err := json.Unmarshal(buf.Bytes(), &check); err != nil {
					t.Errorf("WriteJSON() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestWriteJSON_InvalidType(t *testing.T) {
	// Test with a type that cannot be marshaled to JSON
	buf := &bytes.Buffer{}
	invalidInput := make(chan int) // channels cannot be marshaled to JSON

	err := WriteJSON(buf, invalidInput)
	if err == nil {
		t.Error("WriteJSON() with channel should return error, got nil")
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
		want string
	}{
		{
			name: "text mode",
			mode: Text,
			want: "text",
		},
		{
			name: "json mode",
			mode: JSON,
			want: "json",
		},
		{
			name: "invalid mode defaults to text",
			mode: Mode(99),
			want: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("Mode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModeConstants(t *testing.T) {
	// Verify the constant values are as expected
	if Text != 0 {
		t.Errorf("Text mode value = %v, want 0", Text)
	}
	if JSON != 1 {
		t.Errorf("JSON mode value = %v, want 1", JSON)
	}
}

func TestContextChaining(t *testing.T) {
	// Test that context values don't interfere with each other
	ctx := context.Background()
	ctx = WithMode(ctx, JSON)

	// Add another value to the context
	type otherKey struct{}
	ctx = context.WithValue(ctx, otherKey{}, "other value")

	// Mode should still be retrievable
	if ModeFromContext(ctx) != JSON {
		t.Error("Mode lost after adding another context value")
	}

	// The other value should still be retrievable
	if v := ctx.Value(otherKey{}); v != "other value" {
		t.Error("Other context value lost after setting mode")
	}
}

func TestNilSlicesToEmpty(t *testing.T) {
	t.Run("nil slice in struct becomes empty", func(t *testing.T) {
		type S struct {
			Items []string `json:"items"`
		}
		v := S{Items: nil}
		NilSlicesToEmpty(&v)
		if v.Items == nil {
			t.Fatal("expected non-nil slice after NilSlicesToEmpty")
		}
		if len(v.Items) != 0 {
			t.Fatalf("expected empty slice, got length %d", len(v.Items))
		}
	})

	t.Run("non-nil slice preserved", func(t *testing.T) {
		type S struct {
			Items []string `json:"items"`
		}
		v := S{Items: []string{"a", "b"}}
		NilSlicesToEmpty(&v)
		if len(v.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(v.Items))
		}
	})

	t.Run("nested struct nil slices", func(t *testing.T) {
		type Inner struct {
			Tags []string `json:"tags"`
		}
		type Outer struct {
			Name  string  `json:"name"`
			Inner Inner   `json:"inner"`
			Items []Inner `json:"items"`
		}
		v := Outer{
			Name:  "test",
			Inner: Inner{Tags: nil},
			Items: nil,
		}
		NilSlicesToEmpty(&v)
		if v.Inner.Tags == nil {
			t.Fatal("expected inner tags to be non-nil")
		}
		if v.Items == nil {
			t.Fatal("expected items to be non-nil")
		}
	})

	t.Run("WriteJSON emits empty array not null", func(t *testing.T) {
		type S struct {
			Items []string `json:"items"`
		}
		v := S{Items: nil}
		buf := &bytes.Buffer{}
		if err := WriteJSON(buf, &v); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if strings.Contains(got, "null") {
			t.Errorf("WriteJSON produced null for nil slice: %s", got)
		}
		if !strings.Contains(got, `"items": []`) {
			t.Errorf("WriteJSON did not produce empty array: %s", got)
		}
	})

	t.Run("WriteJSONCompact emits empty array not null", func(t *testing.T) {
		type S struct {
			Items []string `json:"items"`
		}
		v := S{Items: nil}
		buf := &bytes.Buffer{}
		if err := WriteJSONCompact(buf, &v); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if strings.Contains(got, "null") {
			t.Errorf("WriteJSONCompact produced null for nil slice: %s", got)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type S struct {
			Items []int `json:"items"`
		}
		v := &S{Items: nil}
		NilSlicesToEmpty(v)
		if v.Items == nil {
			t.Fatal("expected non-nil slice")
		}
	})

	t.Run("slice of structs with nil slices", func(t *testing.T) {
		type S struct {
			Tags []string `json:"tags"`
		}
		items := []S{{Tags: nil}, {Tags: []string{"a"}}, {Tags: nil}}
		NilSlicesToEmpty(&items)
		for i, item := range items {
			if item.Tags == nil {
				t.Errorf("items[%d].Tags is nil", i)
			}
		}
		if len(items[1].Tags) != 1 || items[1].Tags[0] != "a" {
			t.Error("existing tags were modified")
		}
	})

	t.Run("nil pointer field ignored safely", func(t *testing.T) {
		type Inner struct {
			Tags []string `json:"tags"`
		}
		type Outer struct {
			Ref *Inner `json:"ref"`
		}
		v := Outer{Ref: nil}
		NilSlicesToEmpty(&v) // should not panic
	})
}

func TestConcurrentContextAccess(t *testing.T) {
	// Test that concurrent access to different contexts works correctly
	ctx1 := WithMode(context.Background(), JSON)
	ctx2 := WithMode(context.Background(), Text)

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			if ModeFromContext(ctx1) != JSON {
				t.Error("ctx1 mode changed unexpectedly")
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			if ModeFromContext(ctx2) != Text {
				t.Error("ctx2 mode changed unexpectedly")
			}
		}
		done <- true
	}()

	<-done
	<-done
}
