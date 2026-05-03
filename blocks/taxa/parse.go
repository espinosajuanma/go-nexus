package taxa

import (
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/parser"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Parse(s *scanner.Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		// Blocks end with an END or ENDBLOCK command
		if cmd == "END" || cmd == "ENDBLOCK" {
			return parser.ExpectSemicolon(s)
		}

		switch cmd {
		case "TITLE":
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			if len(tokens) > 0 {
				t.Title = strings.Join(tokens, " ")
			}
		case "DIMENSIONS":
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			for i, tok := range tokens {
				if strings.ToUpper(tok) == "NTAX" {
					valIdx := i + 1
					if valIdx < len(tokens) && tokens[valIdx] == "=" {
						valIdx++
					}
					if valIdx < len(tokens) {
						count, _ := strconv.Atoi(tokens[valIdx])
						t.Dimensions.Count = count
					}
				}
			}
		case "TAXLABELS":
			labels, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			t.TaxLabels = labels
		default:
			if _, err := parser.ReadUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}
