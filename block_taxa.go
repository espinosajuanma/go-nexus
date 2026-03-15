package nexus

import (
	"fmt"
	"strconv"
	"strings"
)

// init automatically registers the TAXA block with the core parser.
func init() {
	RegisterBlock("TAXA", func() Block {
		return &TaxaBlock{}
	})
}

// TaxaBlock specifies information about taxa.
type TaxaBlock struct {
	Dimensions NTax
	TaxLabels  []string
}

type NTax struct {
	Count int
}

// Parse implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Parse(s *Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		// Blocks end with an END or ENDBLOCK command
		if cmd == "END" || cmd == "ENDBLOCK" {
			return expectSemicolon(s)
		}

		switch cmd {
		case "DIMENSIONS":
			tokens, err := readUntilSemicolon(s)
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
			labels, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			t.TaxLabels = labels
		default:
			if _, err := readUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// Render implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Render() string {
	var b strings.Builder
	b.WriteString("BEGIN TAXA;\n")
	b.WriteString(fmt.Sprintf("\tDIMENSIONS NTAX=%d;\n", t.Dimensions.Count))
	b.WriteString("\tTAXLABELS\n")
	for _, label := range t.TaxLabels {
		b.WriteString(fmt.Sprintf("\t\t%s\n", label))
	}
	b.WriteString("\t;\nEND;\n\n")
	return b.String()
}
