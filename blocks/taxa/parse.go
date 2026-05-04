package taxa

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/parser"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Parse(s *scanner.Scanner) error {
	hasDimensions := false
	hasTaxLabels := false

	// Ensure maps are initialized if parsed directly
	if t.TaxSets == nil {
		t.TaxSets = make(map[string]TaxSet)
	}
	if t.TaxPartitions == nil {
		t.TaxPartitions = make(map[string]TaxPartition)
	}

	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		if cmd == "END" || cmd == "ENDBLOCK" {
			return parser.ExpectSemicolon(s)
		}

		switch cmd {
		case "TITLE", "ENTITLE":
			if err := t.parseTitle(s); err != nil {
				return err
			}
		case "DIMENSIONS":
			if err := t.parseDimensions(s, &hasDimensions); err != nil {
				return err
			}
		case "TAXLABELS":
			if err := t.parseTaxLabels(s, hasDimensions, &hasTaxLabels); err != nil {
				return err
			}
		case "TAXSET":
			if err := t.parseTaxSet(s); err != nil {
				return err
			}
		case "TAXPARTITION":
			if err := t.parseTaxPartition(s); err != nil {
				return err
			}
		default:
			if _, err := parser.ReadUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// parseTitle handles the TITLE or ENTITLE commands.
func (t *TaxaBlock) parseTitle(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}
	if len(tokens) > 0 {
		title := strings.Join(tokens, " ")
		t.Title = core.DecodeName(title)
	}
	return nil
}

// parseDimensions handles the DIMENSIONS command and enforces NTAX > 0.
func (t *TaxaBlock) parseDimensions(s *scanner.Scanner, hasDimensions *bool) error {
	if *hasDimensions {
		return fmt.Errorf("DIMENSIONS command can only appear once per TAXA block")
	}
	*hasDimensions = true

	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	ntaxFound := false
	for i, tok := range tokens {
		if strings.ToUpper(tok) == "NTAX" {
			valIdx := i + 1
			if valIdx < len(tokens) && tokens[valIdx] == "=" {
				valIdx++
			}
			if valIdx < len(tokens) {
				count, err := strconv.Atoi(tokens[valIdx])
				if err != nil || count <= 0 {
					return fmt.Errorf("invalid NTAX value: must be a positive integer")
				}
				t.Dimensions = count
				ntaxFound = true
			}
		}
	}

	if !ntaxFound {
		return fmt.Errorf("DIMENSIONS command must include an NTAX parameter")
	}
	return nil
}

// parseTaxLabels populates labels and verifies dimension constraints.
func (t *TaxaBlock) parseTaxLabels(s *scanner.Scanner, hasDimensions bool, hasTaxLabels *bool) error {
	if !hasDimensions {
		return fmt.Errorf("DIMENSIONS must be defined before TAXLABELS")
	}
	if *hasTaxLabels {
		return fmt.Errorf("TAXLABELS command can only appear once per TAXA block")
	}
	*hasTaxLabels = true

	labels, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	if len(labels) != t.Dimensions {
		return fmt.Errorf("dimension mismatch: NTAX declared as %d but found %d TAXLABELS", t.Dimensions, len(labels))
	}

	for _, l := range labels {
		decodedName := core.DecodeName(l)
		normalized := normalizeTaxonName(decodedName)

		if parser.IsAllDigits(normalized) {
			return fmt.Errorf("invalid taxon name '%s': cannot consist entirely of digits", decodedName)
		}

		for _, existing := range t.TaxLabels {
			if normalizeTaxonName(existing) == normalized {
				return fmt.Errorf("duplicate taxon name found during parsing: '%s'", decodedName)
			}
		}

		t.TaxLabels = append(t.TaxLabels, decodedName)
	}
	return nil
}

// parseTaxSet fully implements the AST parsing for TAXSET commands.
func (t *TaxaBlock) parseTaxSet(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}
	if len(tokens) < 3 {
		return fmt.Errorf("malformed TAXSET command")
	}

	name := core.DecodeName(tokens[0])
	format := StandardFormat
	definitionStart := 1

	// Check for format specifiers (e.g., "(VECTOR)")
	if tokens[1] == "(" {
		for i := 2; i < len(tokens); i++ {
			if tokens[i] == ")" {
				definitionStart = i + 1
				break
			}
			if strings.ToUpper(tokens[i]) == "VECTOR" {
				format = VectorFormat
			}
		}
	}

	eqIdx := findEqualsToken(tokens, definitionStart)
	if eqIdx == -1 {
		return fmt.Errorf("expected '=' in TAXSET command")
	}

	t.TaxSets[name] = TaxSet{
		Format:   format,
		TaxaList: tokens[eqIdx+1:], // All tokens after '=' belong to the list
	}
	return nil
}

// parseTaxPartition fully implements the AST parsing for TAXPARTITION commands.
func (t *TaxaBlock) parseTaxPartition(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}
	if len(tokens) < 3 {
		return fmt.Errorf("malformed TAXPARTITION command")
	}

	name := core.DecodeName(tokens[0])
	format := StandardFormat
	definitionStart := 1

	// Check for format specifier
	if tokens[1] == "(" {
		for i := 2; i < len(tokens); i++ {
			if tokens[i] == ")" {
				definitionStart = i + 1
				break
			}
			if strings.ToUpper(tokens[i]) == "VECTOR" {
				format = VectorFormat
			}
		}
	}

	eqIdx := findEqualsToken(tokens, definitionStart)
	if eqIdx == -1 {
		return fmt.Errorf("expected '=' in TAXPARTITION command")
	}

	subsets := make(map[string][]string)
	defTokens := tokens[eqIdx+1:]

	if format == "STANDARD" {
		// Standard format defines explicit subsets with colons: `subset1: 1-3, subset2: 4-6`
		var currentSubset string
		var currentList []string

		for i := 0; i < len(defTokens); i++ {
			tok := defTokens[i]
			if tok == ":" && i > 0 {
				// The previous token was actually the subset name
				currentSubset = core.DecodeName(defTokens[i-1])
				if len(currentList) > 0 {
					currentList = currentList[:len(currentList)-1]
				}
				continue
			}
			if tok == "," {
				if currentSubset != "" {
					subsets[currentSubset] = currentList
				}
				currentSubset = ""
				currentList = nil
				continue
			}
			currentList = append(currentList, tok)
		}
		if currentSubset != "" {
			subsets[currentSubset] = currentList
		}
	} else {
		// Vector format just lists the subset name mapped linearly to taxa by index
		subsets["VECTOR_DATA"] = defTokens
	}

	t.TaxPartitions[name] = TaxPartition{
		Format:  format,
		Subsets: subsets,
	}
	return nil
}

// Helper function to find '=' token safely
func findEqualsToken(tokens []string, startIdx int) int {
	for i := startIdx; i < len(tokens); i++ {
		if tokens[i] == "=" {
			return i
		}
	}
	return -1
}

// converts a taxon name to lowercase for comparisons
func normalizeTaxonName(name string) string {
	return strings.ToLower(name)
}
