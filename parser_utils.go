package nexus

import (
	"fmt"
	"strings"
)

// expectSemicolon is a helper to enforce command termination
func expectSemicolon(s *Scanner) error {
	token, err := s.NextToken()
	if err != nil {
		return err
	}
	if token != ";" {
		return fmt.Errorf("expected ';', got '%s'", token)
	}
	return nil
}

// skipBlock consumes tokens until it finds the END; command.
func skipBlock(s *Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}
		if strings.ToUpper(token) == "END" || strings.ToUpper(token) == "ENDBLOCK" {
			return expectSemicolon(s)
		}
	}
}

func readUntilSemicolon(s *Scanner) ([]string, error) {
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
