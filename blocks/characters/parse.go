package characters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/parser"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Parse(s *scanner.Scanner) error {
	hasDimensions := false

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
			if err := c.parseTitle(s); err != nil {
				return err
			}
		case "DIMENSIONS":
			if hasDimensions {
				return fmt.Errorf("multiple DIMENSIONS commands are not allowed")
			}
			if err := c.parseDimensions(s); err != nil {
				return err
			}
		case "FORMAT":
			if err := c.parseFormat(s); err != nil {
				return err
			}
		case "CHARSTATELABELS":
			if err := c.parseCharStateLabels(s); err != nil {
				return err
			}
		case "MATRIX":
			if err := c.parseMatrix(s); err != nil {
				return err
			}
		case "TAXLABELS":
			// Defines the names of the taxa
			// This command is only permitted in the Characters block if the NewTaxa token was included in the Dimensions command
			continue
		case "CHARLABELS":
			if err := c.parseCharLabels(s); err != nil {
				return err
			}
		case "STATELABELS":
			if err := c.parseStateLabels(s); err != nil {
				return err
			}
		case "ELIMINATE":
			if !hasDimensions {
				return fmt.Errorf("ELIMINATE command must come after DIMENSIONS")
			}
			if err := c.parseEliminate(s); err != nil {
				return err
			}
		default:
			// Skip unrecognized commands
			if _, err := parser.ReadUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// parseTitle reads and sets the block title from a TITLE or ENTITLE command.
func (c *CharactersBlock) parseTitle(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}
	if len(tokens) > 0 {
		c.Title = core.DecodeName(strings.Join(tokens, " "))
	}
	return nil
}

func (c *CharactersBlock) parseDimensions(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	for i, tok := range tokens {
		key := strings.ToUpper(tok)

		// Handle NEWTAXA flag (usually means we are defining taxa here)
		if key == "NEWTAXA" {
			continue
		}

		valIdx := i + 1
		if valIdx < len(tokens) && tokens[valIdx] == "=" {
			valIdx++
		}

		if valIdx < len(tokens) {
			if key == "NCHAR" {
				count, err := strconv.Atoi(tokens[valIdx])
				if err != nil || count <= 0 {
					return fmt.Errorf("invalid NCHAR value: must be a positive integer")
				}
				c.Dimensions = count
				// Pre-populate Character objects with indices
				for j := 1; j <= count; j++ {
					c.Characters = append(c.Characters, &Character{Index: j})
				}
			} else if key == "NTAX" {
				count, err := strconv.Atoi(tokens[valIdx])
				if err != nil || count <= 0 {
					return fmt.Errorf("invalid NTAX value: must be a positive integer")
				}
				// If NTAX is provided, we can pre-allocate the taxa slice if desired
				// or just use it for validation later.
				if len(c.Taxa) == 0 {
					c.Taxa = make([]*TaxonReference, 0, count)
				}
			}
		}
	}
	return nil
}

func (c *CharactersBlock) parseFormat(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}
	// Set standard defaults before parsing
	c.Format.Labels = true

	for i := 0; i < len(tokens); i++ {
		key := strings.ToUpper(tokens[i])

		// Handle boolean flags that don't use an equals sign
		switch key {
		case "INTERLEAVE":
			c.Format.Interleave = true
			continue
		case "TOKENS":
			c.Format.Tokens = true
			continue
		case "RESPECTCASE":
			c.Format.RespectCase = true
			continue
		case "TRANSPOSE":
			c.Format.Transpose = true
			continue
		}

		// Handle key=value pairs
		valIdx := i + 1
		if valIdx < len(tokens) && tokens[valIdx] == "=" {
			valIdx++
			i = valIdx // Skip the '=' token
		}

		if valIdx < len(tokens) {
			val := tokens[valIdx]

			switch key {
			case "DATATYPE":
				dt := DataType(strings.ToUpper(val))
				c.Format.DataType = dt

				// Apply default IUPAC symbols based on the datatype
				if defaultSym, exists := DefaultSymbols[dt]; exists {
					c.Format.Symbols = defaultSym
				}

				// Apply default IUPAC ambiguity equates based on the datatype
				if defaultEq, exists := DefaultEquates[dt]; exists {
					if c.Format.Equate == nil {
						c.Format.Equate = make(map[string]string)
					}
					for k, v := range defaultEq {
						c.Format.Equate[k] = v
					}
				}
			case "MISSING":
				c.Format.Missing = strings.Trim(val, "\"'")
			case "GAP":
				c.Format.Gap = strings.Trim(val, "\"'")
			case "SYMBOLS":
				cleanVal := strings.Trim(val, "\"'")
				var parsedSymbols []string

				// Check if there are spaces. If so, split by space.
				if strings.Contains(cleanVal, " ") {
					parsedSymbols = strings.Fields(cleanVal)
				} else {
					// If smushed (e.g., "012"), split character by character
					for _, ch := range cleanVal {
						parsedSymbols = append(parsedSymbols, string(ch))
					}
				}
				c.Format.Symbols = parsedSymbols
			case "MATCHCHAR":
				c.Format.MatchChar = strings.Trim(val, "\"'")
			case "LABELS":
				if strings.ToUpper(val) == "NO" {
					c.Format.Labels = false
				}
			case "EQUATE":
				if c.Format.Equate == nil {
					c.Format.Equate = make(map[string]string)
				}
				cleanVal := strings.Trim(val, "\"'")
				pairs := strings.Fields(cleanVal) // Split by space
				for _, pair := range pairs {
					parts := strings.SplitN(pair, "=", 2)
					if len(parts) == 2 {
						c.Format.Equate[parts[0]] = parts[1]
					}
				}
			case "NSTATES":
				if n, err := strconv.Atoi(val); err == nil {
					c.Format.NStates = n
				}
			case "ITEMS":
				c.Format.Items = strings.Trim(val, "\"'( )")
			case "STATESFORMAT":
				c.Format.StatesFormat = strings.Trim(val, "\"'")
			}
		}
	}
	return nil
}

func (c *CharactersBlock) parseCharStateLabels(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	var currentID int
	readingStates := false

	for _, t := range tokens {
		// A comma resets the reader for the next character
		if t == "," {
			readingStates = false
			currentID = 0
			continue
		}

		// Catch the character ID
		if id, err := strconv.Atoi(t); err == nil && !readingStates {
			currentID = id
			continue
		}

		// Toggle state-reading mode when we hit a slash
		if t == "/" {
			readingStates = true
			continue
		}

		// Apply names and states to the specific Character object
		if currentID != 0 && currentID <= len(c.Characters) {
			char := c.Characters[currentID-1]

			if !readingStates {
				char.Name = core.DecodeName(t)
			} else {
				stateName := core.DecodeName(t)
				if stateName == "_" {
					stateName = "" // Translate NEXUS underscores back into empty strings internally
				}
				char.StateLabels = append(char.StateLabels, stateName)
			}
		}
	}
	return nil
}

func (c *CharactersBlock) parseMatrix(s *scanner.Scanner) error {
	nchar := c.Dimensions
	progress := make(map[string]int) // Tracks char index per normalized taxon name
	var firstTaxonName string        // Tracks the normalized name of the first taxon for MATCHCHAR

	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}
		if token == ";" {
			break
		}

		// Resolve Taxon Name (Normalized for logic, Raw for UI)
		normalizedName := normalizeTaxonName(token)
		var taxonRef *TaxonReference

		// Check for existing taxon (Case-insensitive Interleave support)
		exists := false
		for _, t := range c.Taxa {
			if normalizeTaxonName(t.Name) == normalizedName {
				taxonRef = t
				exists = true
				break
			}
		}

		// Register new taxon if not seen before
		if !exists {
			// Pass the raw token to AddTaxon so the original case is preserved for rendering
			taxonRef = c.AddTaxon(token)
		}
		if firstTaxonName == "" {
			firstTaxonName = normalizedName
		}

		// Parse Characters for this segment
		for progress[normalizedName] < nchar {
			stateTok, err := s.NextToken()
			if err != nil {
				return err
			}

			// If we hit a semicolon prematurely
			if stateTok == ";" {
				return fmt.Errorf("unexpected ';' in matrix for taxon %s", taxonRef.Name)
			}

			// Decompose the token based on FORMAT (TOKENS vs Smushed)
			states, err := c.decomposeStateToken(stateTok, s)
			if err != nil {
				return err
			}

			for _, cs := range states {
				currIdx := progress[normalizedName]
				if currIdx >= nchar {
					break
				}

				// Handle MATCHCHAR logic safely using the first taxon
				if len(cs.Value) == 1 && cs.Value[0] == c.Format.MatchChar && normalizedName != firstTaxonName {
					// Copy state from the corresponding position in the first taxon
					cs = c.data[0][currIdx]
				} else {
					// Map literal symbols to internal sentinels
					cs.Value = c.resolveSpecialSymbols(cs.Value)
					// Apply EQUATE macros
					cs.Value = c.applyEquates(cs.Value)
				}

				c.data[taxonRef.Index][currIdx] = cs
				progress[normalizedName]++
			}

			// If INTERLEAVE is true, we stop after one line/segment.
			// Heuristic: if the next token is a known taxon, we break this segment's loop.
			if c.Format.Interleave {
				peek, _ := s.PeekToken()
				if c.isKnownTaxon(peek) || peek == ";" {
					break
				}
			}
		}
	}
	return nil
}

// isKnownTaxon checks if a token matches a taxon name already registered in this block.
func (c *CharactersBlock) isKnownTaxon(name string) bool {
	normalized := normalizeTaxonName(name)
	for _, t := range c.Taxa {
		if normalizeTaxonName(t.Name) == normalized {
			return true
		}
	}
	return false
}

// decomposeStateToken breaks a token into CharacterState objects based on FORMAT rules.
func (c *CharactersBlock) decomposeStateToken(tok string, s *scanner.Scanner) ([]CharacterState, error) {
	// Handle bracketed groups: (polymorphic) or {uncertain}
	if tok == "(" || tok == "{" {
		stateType := StatePolymorphic
		closing := ")"
		if tok == "{" {
			stateType = StateUncertain
			closing = "}"
		}

		var subValues []string
		for {
			inner, err := s.NextToken()
			if err != nil {
				return nil, err
			}
			if inner == closing {
				break
			}
			subValues = append(subValues, inner)
		}
		return []CharacterState{{Type: stateType, Value: subValues}}, nil
	}

	// Handle standard tokens: if FORMAT TOKENS is set, the whole word is ONE state.
	if c.Format.Tokens {
		return []CharacterState{{Type: StateSingle, Value: []string{tok}}}, nil
	}

	// Handle smushed sequences: "ATGC" -> A, T, G, C
	var states []CharacterState
	for _, char := range tok {
		states = append(states, CharacterState{
			Type:  StateSingle,
			Value: []string{string(char)},
		})
	}
	return states, nil
}

// resolveSpecialSymbols converts literal '?' or '-' into internal sentinels.
func (c *CharactersBlock) resolveSpecialSymbols(vals []string) []string {
	resolved := make([]string, len(vals))
	for i, v := range vals {
		switch v {
		case c.Format.Missing:
			resolved[i] = InternalMissing
		case c.Format.Gap:
			resolved[i] = InternalGap
		default:
			resolved[i] = v
		}
	}
	return resolved
}

// applyEquates resolves custom EQUATE mappings (e.g., "R" -> "A G").
func (c *CharactersBlock) applyEquates(vals []string) []string {
	var final []string
	for _, v := range vals {
		if mapping, ok := c.Format.Equate[v]; ok {
			// EQUATE values can be multi-state like "(A G)"
			clean := strings.Trim(mapping, "() ")
			final = append(final, strings.Fields(clean)...)
		} else {
			final = append(final, v)
		}
	}
	return final
}

// parseCharLabels assigns names to existing Character objects sequentially.
// Syntax: CHARLABELS eye_color hair_color ... ;
func (c *CharactersBlock) parseCharLabels(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	charIndex := 0
	for _, t := range tokens {
		if t == "," {
			continue // Commas are optional separators
		}
		if charIndex < len(c.Characters) {
			c.Characters[charIndex].Name = core.DecodeName(t)
			charIndex++
		}
	}
	return nil
}

// parseStateLabels applies states to specific characters by ID.
// Syntax: STATELABELS 1 A G, 2 red blue ;
func (c *CharactersBlock) parseStateLabels(s *scanner.Scanner) error {
	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	var currentID int

	for _, t := range tokens {
		if t == "," {
			currentID = 0 // Reset for the next character mapping
			continue
		}

		// If we don't have a current ID, this token should be an integer ID
		if currentID == 0 {
			if id, err := strconv.Atoi(t); err == nil {
				currentID = id
			}
			continue
		}

		// Otherwise, it's a state label for the current character ID
		if currentID <= len(c.Characters) {
			stateName := core.DecodeName(t)
			if stateName == "_" {
				stateName = ""
			}
			c.Characters[currentID-1].StateLabels = append(c.Characters[currentID-1].StateLabels, stateName)
		}
	}
	return nil
}

// parseEliminate tracks which characters should be ignored/dropped.
// Syntax: ELIMINATE 4-10, 15;
func (c *CharactersBlock) parseEliminate(s *scanner.Scanner) error {
	if c.Eliminate == nil {
		c.Eliminate = make(map[int]bool)
	}

	tokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	// Re-join tokens and split by comma to handle ranges properly
	joined := strings.Join(tokens, "")
	parts := strings.Split(joined, ",")

	for _, part := range parts {
		if strings.Contains(part, "-") {
			// Handle Range (e.g. 4-10)
			bounds := strings.Split(part, "-")
			if len(bounds) == 2 {
				start, err1 := strconv.Atoi(bounds[0])
				end, err2 := strconv.Atoi(bounds[1])
				if err1 == nil && err2 == nil {
					for i := start; i <= end; i++ {
						c.Eliminate[i] = true
					}
				}
			}
		} else {
			// Handle Single integer
			if val, err := strconv.Atoi(part); err == nil {
				c.Eliminate[val] = true
			}
		}
	}
	return nil
}

// normalizeTaxonName resolves NEXUS string rules and enforces case-insensitivity
func normalizeTaxonName(name string) string {
	return strings.ToLower(core.DecodeName(name))
}
