package characters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/espinosajuanma/go-nexus/parser"
	"github.com/espinosajuanma/go-nexus/scanner"
	"github.com/espinosajuanma/go-nexus/utils"
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
		c.Title = utils.DecodeName(strings.Join(tokens, " "))
	}
	return nil
}

// parseDimensions reads the DIMENSIONS command to determine the number of characters and optionally taxa.
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
				for j := 1; j <= count; j++ {
					c.Matrix.Characters = append(c.Matrix.Characters, &Character{Index: j})
				}
			} else if key == "NTAX" {
				count, err := strconv.Atoi(tokens[valIdx])
				if err != nil || count <= 0 {
					return fmt.Errorf("invalid NTAX value: must be a positive integer")
				}
				// If NTAX is provided, we can pre-allocate the taxa slice if desired
				// or just use it for validation later.
				if len(c.Matrix.Taxa) == 0 {
					c.Matrix.Taxa = make([]*Taxon, 0, count)
				}
			}
		}
	}
	return nil
}

// parseFormat reads the FORMAT command and sets parsing rules accordingly.
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

// parseCharStateLabels assigns names to specific character states by character ID.
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
		if currentID != 0 && currentID <= len(c.Matrix.Characters) {
			char := c.Matrix.Characters[currentID-1]

			if !readingStates {
				char.Name = utils.DecodeName(t)
			} else {
				stateName := utils.DecodeName(t)
				if stateName == "_" {
					stateName = "" // Translate NEXUS underscores back into empty strings internally
				}
				char.StateLabels = append(char.StateLabels, stateName)
			}
		}
	}
	return nil
}

// parseMatrix reads the MATRIX command and fills the data matrix.
func (c *CharactersBlock) parseMatrix(s *scanner.Scanner) error {
	nchar := c.Dimensions

	// Track how many character states we've read for each taxon to know when we're done with that row.
	progress := make(map[int]int)

	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}
		if token == ";" {
			break
		}

		taxon := c.Matrix.GetTaxon(token)
		if taxon == nil {
			taxon = c.AddTaxon(token)
		}

		for progress[taxon.Index] < nchar {
			stateTok, err := s.NextToken()
			if err != nil {
				return err
			}

			if stateTok == ";" {
				return fmt.Errorf("unexpected ';' in matrix for taxon %s", taxon.Name)
			}

			states, err := c.decomposeStateToken(stateTok, s)
			if err != nil {
				return err
			}

			for _, cs := range states {
				currIdx := progress[taxon.Index]
				if currIdx >= nchar {
					break
				}

				// --- MATCHCHAR LOGIC ---
				if len(cs.Values) == 1 && cs.Values[0].Symbol == c.Format.MatchChar {
					if taxon.Index > 0 {
						// Deep copy state from taxon 0 at the exact same column
						firstTaxonState := c.Matrix.GetStateByIndex(0, currIdx)
						cs.Type = firstTaxonState.Type
						// Safely copy the slice so mutations don't bleed across rows
						cs.Values = append([]StateValue(nil), firstTaxonState.Values...)
						cs.Values = c.applyEquates(cs.Values) // Apply equates to the copied state as well, in case the first taxon had an equate symbol. This ensures matchchar also respects equates.
					}
				} else {
					// Apply equates to newly parsed values
					cs.Values = c.applyEquates(cs.Values)
				}

				c.Matrix.SetStateByIndex(taxon.Index, currIdx, cs)
				progress[taxon.Index]++
			}

			if c.Format.Interleave {
				peek, _ := s.PeekToken()
				// Break if the states are smushed (1 token = 1 chunk),
				// OR if the next token is a taxon we recognize,
				// OR if we hit the end of the block.
				if !c.Format.Tokens || c.Matrix.GetTaxon(peek) != nil || peek == ";" {
					break
				}
			}
		}
	}
	return nil
}

// parseValue checks for the ":" separator used in COUNT and FREQUENCY formats.
func parseValue(token string) StateValue {
	parts := strings.SplitN(token, ":", 2)
	value := StateValue{Symbol: parts[0], Weight: 1.0}

	if len(parts) == 2 {
		if weight, err := strconv.ParseFloat(parts[1], 64); err == nil {
			value.Weight = weight
		}
	}
	return value
}

// expandRange takes a token like "0~3" and returns the intermediate states
// based on the defined Format.Symbols list.
func (c *CharactersBlock) expandRange(token string) ([]StateValue, error) {
	if !strings.Contains(token, "~") {
		return []StateValue{parseValue(token)}, nil
	}

	parts := strings.Split(token, "~")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", token)
	}

	startSym, endSym := parseValue(parts[0]), parseValue(parts[1])

	startIndex, endIndex := -1, -1
	for i, sym := range c.Format.Symbols {
		if sym == startSym.Symbol {
			startIndex = i
		}
		if sym == endSym.Symbol {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 || startIndex > endIndex {
		return nil, fmt.Errorf("invalid or out-of-order range symbols: %s", token)
	}

	var expanded []StateValue
	for i := startIndex; i <= endIndex; i++ {
		// Assigning the start weight to all expanded items (or default 1.0)
		expanded = append(expanded, StateValue{
			Symbol: c.Format.Symbols[i],
			Weight: startSym.Weight,
		})
	}
	return expanded, nil
}

// isKnownTaxon checks if a token matches a taxon name already registered in this block.
func (c *CharactersBlock) isKnownTaxon(name string) bool {
	return c.Matrix.GetTaxon(name) != nil
}

// decomposeStateToken breaks a matrix token into one or more robust CharacterState objects.
// It returns a slice because a single token (e.g., "ATGC") may represent multiple character columns if TOKENS is false.
func (c *CharactersBlock) decomposeStateToken(token string, s *scanner.Scanner) ([]CharacterState, error) {
	var states []CharacterState

	// Handle Polymorphic () or Uncertain {} blocks
	if token == "(" || token == "{" {
		state := CharacterState{Type: StatePolymorphic}
		closing := ")"
		if token == "{" {
			state.Type = StateUncertain
			closing = "}"
		}

		for {
			inner, err := s.NextToken()
			if err != nil {
				return nil, err
			}
			if inner == closing {
				break
			}

			// Smushed strings inside parenthesis (e.g., "(AC)")
			if !c.Format.Tokens && !strings.Contains(inner, "~") && len(inner) > 1 {
				for _, ch := range inner {
					state.Values = append(state.Values, parseValue(string(ch)))
				}
			} else {
				// Expand ranges (e.g., "0~3") and append to values
				vals, err := c.expandRange(inner)
				if err != nil {
					return nil, err
				}
				state.Values = append(state.Values, vals...)
			}
		}
		// A polymorphic/uncertain block always applies to a SINGLE character column
		return []CharacterState{state}, nil
	}

	// Handle literal Missing or Gap symbols (if spaced out)
	if token == c.Format.Missing {
		return []CharacterState{{Type: StateMissing}}, nil
	}
	if token == c.Format.Gap {
		return []CharacterState{{Type: StateGap}}, nil
	}

	// Handle Single States (Tokens vs Smushed)
	if c.Format.Tokens {
		// If TOKENS is set, the whole word belongs to ONE character column (e.g., "absent")
		state := CharacterState{
			Type:   StateSingle,
			Values: []StateValue{parseValue(token)},
		}
		return []CharacterState{state}, nil
	}

	// Handle Smushed Sequences (e.g., "ATGC" -> A, T, G, C)
	// If TOKENS is false, we must break the token character-by-character.
	for _, ch := range token {
		charStr := string(ch)

		switch charStr {
		case c.Format.Missing:
			states = append(states, CharacterState{Type: StateMissing})
		case c.Format.Gap:
			states = append(states, CharacterState{Type: StateGap})
		case c.Format.MatchChar:
			// Catch match characters inside smushed sequences (e.g., "A..C")
			states = append(states, CharacterState{
				Type:   StateSingle,
				Values: []StateValue{{Symbol: charStr, Weight: 1.0}},
			})
		default:
			states = append(states, CharacterState{
				Type:   StateSingle,
				Values: []StateValue{parseValue(charStr)},
			})
		}
	}

	return states, nil
}

// applyEquates maps a symbol to its expanded values (e.g. "R" -> "A" and "G")
func (c *CharactersBlock) applyEquates(vals []StateValue) []StateValue {
	var final []StateValue
	for _, o := range vals {
		if mapping, ok := c.Format.Equate[o.Symbol]; ok {
			clean := strings.Trim(mapping, "() ")
			fields := strings.FieldsSeq(clean)
			for f := range fields {
				final = append(final, StateValue{
					Symbol: f,
					Weight: o.Weight, // Preserve weight across the expanded states
				})
			}
		} else {
			final = append(final, o)
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
		if charIndex < len(c.Matrix.Characters) {
			c.Matrix.Characters[charIndex].Name = utils.DecodeName(t)
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
		if currentID <= len(c.Matrix.Characters) {
			stateName := utils.DecodeName(t)
			if stateName == "_" {
				stateName = ""
			}
			c.Matrix.Characters[currentID-1].StateLabels = append(c.Matrix.Characters[currentID-1].StateLabels, stateName)
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
	return strings.ToLower(utils.DecodeName(name))
}
