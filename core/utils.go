package core

import "strings"

// DecodeName normalizes NEXUS unquoted identifiers by replacing underscores with spaces.
func DecodeName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), "_", " ")
}

// EncodeName replaces spaces with underscores for safe, unquoted NEXUS rendering.
// Best used for Taxa labels and Matrix row headers to maintain alignment.
func EncodeName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}

// QuoteName conditionally wraps a string in single quotes if it contains spaces
// or NEXUS punctuation, escaping any internal single quotes by doubling them.
// Best used for Character Names, State Labels, and Titles.
func QuoteName(name string) string {
	name = strings.TrimSpace(name)
	// If it's a completely safe word, return as is.
	if !strings.ContainsAny(name, " \t\n(){}[]/\\,;:=*\"+-<>~") && !strings.Contains(name, "'") {
		return name
	}
	// Escape existing single quotes by doubling them (' -> '')
	escaped := strings.ReplaceAll(name, "'", "''")
	return "'" + escaped + "'"
}
