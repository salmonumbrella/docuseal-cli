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
// Nil slices are normalized to empty arrays to avoid null in output.
func WriteJSON(w io.Writer, v any) error {
	NilSlicesToEmpty(v)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteJSONCompact writes a value as compact JSON (one object/array).
// Nil slices are normalized to empty arrays to avoid null in output.
func WriteJSONCompact(w io.Writer, v any) error {
	NilSlicesToEmpty(v)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// WriteNDJSON writes a slice/array value as newline-delimited JSON (one element per line).
// If v is not a slice/array, it falls back to compact JSON encoding of v.
// Nil slices are normalized to empty arrays to avoid null in output.
func WriteNDJSON(w io.Writer, v any) error {
	NilSlicesToEmpty(v)
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

// NilSlicesToEmpty recursively walks a value and replaces nil slices with
// empty slices so that JSON marshaling produces [] instead of null.
// It modifies the value in place via reflection and only operates on
// settable fields (structs, slices, pointers, maps).
func NilSlicesToEmpty(v any) {
	nilSlicesWalk(reflect.ValueOf(v), make(map[uintptr]bool))
}

func nilSlicesWalk(rv reflect.Value, visited map[uintptr]bool) {
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return
		}
		// Guard against pointer cycles.
		ptr := rv.Pointer()
		if visited[ptr] {
			return
		}
		visited[ptr] = true
		nilSlicesWalk(rv.Elem(), visited)

	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Field(i)
			if !f.CanSet() {
				continue
			}
			nilSlicesWalk(f, visited)
		}

	case reflect.Slice:
		if rv.IsNil() && rv.CanSet() {
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))
			return
		}
		for i := 0; i < rv.Len(); i++ {
			nilSlicesWalk(rv.Index(i), visited)
		}

	case reflect.Map:
		if rv.IsNil() {
			return
		}
		for _, key := range rv.MapKeys() {
			elem := rv.MapIndex(key)
			// Map values are not directly settable; if the value is a struct
			// or contains slices we need to copy, modify, and re-set.
			if needsWalk(elem) {
				cp := reflect.New(elem.Type()).Elem()
				cp.Set(elem)
				nilSlicesWalk(cp, visited)
				rv.SetMapIndex(key, cp)
			}
		}

	case reflect.Interface:
		if rv.IsNil() {
			return
		}
		elem := rv.Elem()
		if needsWalk(elem) {
			// Interface values are not directly settable; unwrap, walk, re-set.
			cp := reflect.New(elem.Type()).Elem()
			cp.Set(elem)
			nilSlicesWalk(cp, visited)
			if rv.CanSet() {
				rv.Set(cp)
			}
		}
	}
}

// needsWalk returns true if the value could contain nil slices.
func needsWalk(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Ptr, reflect.Struct, reflect.Slice, reflect.Map, reflect.Interface:
		return true
	}
	return false
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
