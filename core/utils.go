package core

import "strings"

// SanitizeName replaces spaces with underscores to conform to NEXUS word rules.
func SanitizeName(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), " ", "_")
}
