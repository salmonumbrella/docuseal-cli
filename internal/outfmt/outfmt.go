package outfmt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// Mode represents the output format mode
type Mode int

const (
	Text Mode = iota
	JSON
	NDJSON
)

type contextKey struct{}

// Parse parses an output mode string
func Parse(s string) (Mode, error) {
	switch s {
	case "text", "":
		return Text, nil
	case "json":
		return JSON, nil
	case "ndjson", "jsonl":
		return NDJSON, nil
	default:
		return Text, fmt.Errorf("invalid output format: %q (use 'text', 'json', or 'ndjson')", s)
	}
}

// WithMode adds the output mode to the context
func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, contextKey{}, mode)
}

// ModeFromContext retrieves the output mode from context
func ModeFromContext(ctx context.Context) Mode {
	if mode, ok := ctx.Value(contextKey{}).(Mode); ok {
		return mode
	}
	return Text
}

// IsJSON returns true if the context is set to JSON output
func IsJSON(ctx context.Context) bool {
	mode := ModeFromContext(ctx)
	return mode == JSON || mode == NDJSON
}

// WriteJSON writes a value as pretty-printed JSON.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteJSONCompact writes a value as compact JSON (one object/array).
func WriteJSONCompact(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// WriteNDJSON writes a slice/array value as newline-delimited JSON (one element per line).
// If v is not a slice/array, it falls back to compact JSON encoding of v.
func WriteNDJSON(w io.Writer, v any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return WriteJSONCompact(w, v)
	}

	kind := rv.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return WriteJSONCompact(w, v)
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for i := 0; i < rv.Len(); i++ {
		if err := enc.Encode(rv.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

// String returns the string representation of the mode
func (m Mode) String() string {
	switch m {
	case JSON:
		return "json"
	case NDJSON:
		return "ndjson"
	default:
		return "text"
	}
}
