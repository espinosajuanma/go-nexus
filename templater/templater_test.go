package templater

import (
	"reflect"
	"strings"
	"testing"
	"text/template"
)

// TestHelperFunctions directly tests the unexported string manipulation functions.
func TestHelperFunctions(t *testing.T) {
	t.Run("snake", func(t *testing.T) {
		tests := []struct{ input, expected string }{
			{"hello world", "hello_world"},
			{"no_spaces", "no_spaces"},
			{" leading space", "_leading_space"},
		}
		for _, tt := range tests {
			if got := snake(tt.input); got != tt.expected {
				t.Errorf("snake(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("quote", func(t *testing.T) {
		tests := []struct{ input, expected string }{
			{"simple", "simple"},                       // No spaces or special chars
			{"has spaces", "'has spaces'"},             // Spaces trigger quotes
			{"has'quote", "'has''quote'"},              // Single quotes are escaped
			{"already''escaped", "'already''escaped'"}, // Double single-quotes normalized then re-escaped
			{"has\"double", "'has\"double'"},           // Double quotes trigger outer single quotes
		}
		for _, tt := range tests {
			if got := quote(tt.input); got != tt.expected {
				t.Errorf("quote(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("pad", func(t *testing.T) {
		// pad adds width + 2. If width is 5, total space is 7.
		got := pad(5, "foo")
		expected := "foo    " // 3 chars + 4 spaces = 7
		if got != expected {
			t.Errorf("pad(5, 'foo') = %q (len %d); want %q (len %d)", got, len(got), expected, len(expected))
		}
	})

	t.Run("nullName", func(t *testing.T) {
		if got := nullName(""); got != "_" {
			t.Errorf("nullName(\"\") = %q; want \"_\"", got)
		}
		if got := nullName("simple"); got != "simple" {
			t.Errorf("nullName(\"simple\") = %q; want \"simple\"", got)
		}
		if got := nullName("has space"); got != "'has space'" {
			t.Errorf("nullName(\"has space\") = %q; want \"'has space'\"", got)
		}
	})

	t.Run("wrap", func(t *testing.T) {
		if got := wrap("[", "]", "content"); got != "[content]" {
			t.Errorf("wrap() = %q; want \"[content]\"", got)
		}
		if got := wrap("[", "]", ""); got != "" {
			t.Errorf("wrap() on empty string = %q; want \"\"", got)
		}
	})

	t.Run("join", func(t *testing.T) {
		got := join(", ", []string{"a", "b", "c"})
		if got != "a, b, c" {
			t.Errorf("join() = %q; want \"a, b, c\"", got)
		}
	})

	t.Run("sortMap", func(t *testing.T) {
		m := map[string]string{"c": "3", "a": "1", "b": "2"}
		got := sortMap(m)
		expected := []Pair{
			{Key: "a", Value: "1"},
			{Key: "b", Value: "2"},
			{Key: "c", Value: "3"},
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("sortMap() = %v; want %v", got, expected)
		}
	})
}

// TestTemplateRendering verifies the exported API and template execution.
func TestTemplateRendering(t *testing.T) {
	t.Run("RenderString Success", func(t *testing.T) {
		layout := `Name: {{.Name | snake | quote}}`
		data := map[string]string{"Name": "my bad name"}

		got, err := RenderString(layout, data)
		if err != nil {
			t.Fatalf("RenderString returned unexpected error: %v", err)
		}

		expected := `Name: my_bad_name` // 'snake' removes the space, so 'quote' doesn't wrap it!
		if got != expected {
			t.Errorf("RenderString() = %q; want %q", got, expected)
		}
	})

	t.Run("RenderString Invalid Template", func(t *testing.T) {
		_, err := RenderString(`{{.Unclosed`, nil)
		if err == nil {
			t.Error("Expected RenderString to fail on invalid template syntax")
		}
	})

	t.Run("New with Custom FuncMap", func(t *testing.T) {
		layout := `{{.Val | customFunc}}`

		// Create a custom function map to pass into New
		custom := template.FuncMap{
			"customFunc": func(s string) string {
				return strings.ToUpper(s)
			},
		}

		tmpl, err := New("test", layout, custom)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}

		data := map[string]string{"Val": "lower"}
		got, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Render() failed: %v", err)
		}

		if got != "LOWER" {
			t.Errorf("Render() with custom func = %q; want %q", got, "LOWER")
		}
	})

	t.Run("Render Execution Error", func(t *testing.T) {
		// A valid template, but executing an invalid field access to trigger an execution error
		tmpl, _ := New("test", `{{.Missing.Field}}`)

		// Passing a string instead of a struct will cause an execution panic/error
		_, err := tmpl.Render("not a struct")
		if err == nil {
			t.Error("Expected Render() to return an error when data doesn't match template execution")
		}
	})
}
