package nexus

import (
	"fmt"
	"strconv"
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

func parseFormatCommand(tokens []string) Format {
	format := Format{Missing: "?"}
	for i := 0; i < len(tokens); i++ {
		key := strings.ToUpper(tokens[i])

		valIdx := i + 1
		if valIdx < len(tokens) && tokens[valIdx] == "=" {
			valIdx++
			i = valIdx
		}

		if valIdx < len(tokens) {
			val := tokens[valIdx]
			switch key {
			case "DATATYPE":
				format.DataType = strings.ToUpper(val)
			case "MISSING":
				format.Missing = val
			case "GAP":
				format.Gap = val
			case "SYMBOLS":
				format.Symbols = strings.Trim(val, "\"'")
			}
		}
	}
	return format
}

func parseCharStateLabels(s *Scanner, labels map[int]string) error {
	tokens, err := readUntilSemicolon(s)
	if err != nil {
		return err
	}

	for i := 0; i < len(tokens); i++ {
		num, err := strconv.Atoi(tokens[i])
		if err == nil && i+1 < len(tokens) {
			labels[num] = tokens[i+1]
		}
	}
	return nil
}

func parseMatrix(s *Scanner, chars *CharactersBlock) error {
	for {
		taxonToken, err := s.NextToken()
		if err != nil {
			return err
		}
		if taxonToken == ";" {
			break
		}

		dataToken, err := s.NextToken()
		if err != nil {
			return fmt.Errorf("expected data for taxon %s: %w", taxonToken, err)
		}

		chars.Matrix = append(chars.Matrix, MatrixRow{
			TaxonName: taxonToken,
			Data:      dataToken,
		})
	}
	return nil
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
