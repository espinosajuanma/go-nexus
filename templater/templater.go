package templater

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

// Template provides a simple wrapper around text/template with some Nexus-specific helper functions.
type Template struct {
	inner *template.Template
}

// New creates a new Template instance by parsing the provided layout string.
func New(name, layout string) (*Template, error) {
	fns := template.FuncMap{
		"snake":    snake,
		"quote":    quote,
		"pad":      pad,
		"nullName": nullName,
		"wrap":     wrap,
		"join":     join,
		"sortMap":  sortMap,
	}

	t, err := template.New(name).Funcs(fns).Parse(layout)
	if err != nil {
		return nil, err
	}
	return &Template{inner: t}, nil
}

// Render executes the template with the given data and returns the resulting string.
func (t *Template) Render(data any) (string, error) {
	var buf bytes.Buffer
	if err := t.inner.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderString is a convenience function that creates a Template and renders it in one step.
func RenderString(layout string, data any) (string, error) {
	t, err := New("one-off", layout)
	if err != nil {
		return "", err
	}
	return t.Render(data)
}

// snake converts a string to snake_case by replacing spaces with underscores.
func snake(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

// quote wraps a string in double quotes and escapes any existing double quotes.
func quote(s string) string {
	normalized := strings.ReplaceAll(s, "''", "'")
	escaped := strings.ReplaceAll(normalized, "'", "''")
	return "'" + escaped + "'"
}

// pad adds spaces to the right of a string to ensure it reaches a specified width.
func pad(width int, s string) string {
	return fmt.Sprintf("%-*s", width+2, s)
}

// nullName returns an underscore if the input string is empty, otherwise it returns the original string.
func nullName(s string) string {
	if s == "" {
		return "_"
	}
	return quote(s)
}

// wrap conditionally wraps content with specified start and end strings.
func wrap(start, end, content string) string {
	if content == "" {
		return ""
	}
	return start + content + end
}

// join concatenates a slice of strings using a specified separator.
func join(sep string, items []string) string {
	return strings.Join(items, sep)
}

// Pair is a simple struct to hold key-value pairs for sorted maps.
type Pair struct {
	Key   string
	Value any
}

// sortMap takes a map and returns a sorted slice of key-value pairs based on the keys.
func sortMap(m map[string]string) []Pair {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]Pair, len(keys))
	for i, k := range keys {
		pairs[i] = Pair{Key: k, Value: m[k]}
	}
	return pairs
}
