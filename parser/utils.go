package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/espinosajuanma/nexus/scanner"
)

// ExpectSemicolon is a helper to enforce command termination
func ExpectSemicolon(s *scanner.Scanner) error {
	token, err := s.NextToken()
	if err != nil {
		return err
	}
	if token != ";" {
		return fmt.Errorf("expected ';', got '%s'", token)
	}
	return nil
}

// SkipBlock consumes tokens until it finds the END; command.
func SkipBlock(s *scanner.Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}
		if strings.ToUpper(token) == "END" || strings.ToUpper(token) == "ENDBLOCK" {
			return ExpectSemicolon(s)
		}
	}
}

// ReadUntilSemicolon reads tokens until it finds the END; command.
func ReadUntilSemicolon(s *scanner.Scanner) ([]string, error) {
	var tokens []string
	for {
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == ";" {
			break
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// IsAllDigits checks if a string consists entirely of digits (0-9)
func IsAllDigits(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}
