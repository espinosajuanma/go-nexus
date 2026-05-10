package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Scanner wraps a bufio.Reader to tokenize a NEXUS file.
type Scanner struct {
	reader      *bufio.Reader
	peekedToken *string
	peekedErr   error
}

// NewScanner creates a new NEXUS scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{reader: bufio.NewReader(r)}
}

// NextToken returns the next NEXUS token. It skips whitespace and comments.
// Returns an empty string and io.EOF when the end of the file is reached.
func (s *Scanner) NextToken() (string, error) {
	// Consume the peeked token if one exists
	if s.peekedToken != nil {
		tok := *s.peekedToken
		err := s.peekedErr

		// Reset peek state
		s.peekedToken = nil
		s.peekedErr = nil

		return tok, err
	}

	// Otherwise, read a fresh token using the helper
	return s.readNextToken()
}

// PeekToken looks at the next token without advancing the scanner's position.
func (s *Scanner) PeekToken() (string, error) {
	if s.peekedToken != nil {
		return *s.peekedToken, s.peekedErr
	}

	tok, err := s.readNextToken()
	s.peekedToken = &tok
	s.peekedErr = err
	return tok, err
}

// ReadRawUntilBlockEnd reads all characters until it reaches END; or ENDBLOCK;
// It returns the raw string content excluding the END; command.
func (s *Scanner) ReadRawUntilBlockEnd() (string, error) {
	var content bytes.Buffer
	var token bytes.Buffer

	if s.peekedToken != nil {
		content.WriteString(*s.peekedToken)
		token.WriteString(*s.peekedToken)
		s.peekedToken = nil
		s.peekedErr = nil
	}

	inComment := 0
	inQuote := false

	for {
		ch, _, err := s.reader.ReadRune()
		if err != nil {
			return "", err
		}
		content.WriteRune(ch)

		// Handle Quotes
		if ch == '\'' {
			inQuote = !inQuote
			token.Reset()
			continue
		}
		if inQuote {
			continue
		}

		// Handle Nested Comments
		if ch == '[' {
			inComment++
			token.Reset()
			continue
		}
		if ch == ']' {
			if inComment > 0 {
				inComment--
			}
			token.Reset()
			continue
		}
		if inComment > 0 {
			continue
		}

		// Handle Spaces
		if unicode.IsSpace(ch) {
			cmd := strings.ToUpper(token.String())
			// Don't reset the token if it's "END", so we can catch instances like "END  ;"
			if cmd != "END" && cmd != "ENDBLOCK" {
				token.Reset()
			}
			continue
		}

		// Handle Semicolons and Commands
		if isPunctuation(ch) {
			if ch == ';' {
				cmd := strings.ToUpper(token.String())
				if cmd == "END" || cmd == "ENDBLOCK" {
					full := content.String()
					// Strip out the END; / ENDBLOCK; part from the string
					upperFull := strings.ToUpper(full)
					idx := strings.LastIndex(upperFull, cmd)
					if idx != -1 {
						return full[:idx], nil
					}
					return full, nil
				}
			}
			token.Reset()
			continue
		}

		token.WriteRune(ch)
	}
}

// readNextToken contains the core tokenization logic.
func (s *Scanner) readNextToken() (string, error) {
	for {
		ch, _, err := s.reader.ReadRune()
		if err != nil {
			return "", err
		}

		if unicode.IsSpace(ch) {
			continue
		}
		if ch == '[' {
			if err := s.skipComment(); err != nil {
				return "", err
			}
			continue
		}
		if ch == '\'' {
			return s.readQuotedWord()
		}
		if isPunctuation(ch) {
			return string(ch), nil
		}

		s.reader.UnreadRune()
		return s.readWord()
	}
}

// skipComment handles nested comments by counting open/close brackets.
func (s *Scanner) skipComment() error {
	depth := 1
	for depth > 0 {
		ch, _, err := s.reader.ReadRune()
		if err != nil {
			return fmt.Errorf("unexpected EOF inside comment: %w", err)
		}
		switch ch {
		case '[':
			depth++
		case ']':
			depth--
		}
	}
	return nil
}

// readQuotedWord reads a word bounded by single quotes.
func (s *Scanner) readQuotedWord() (string, error) {
	var buf bytes.Buffer
	for {
		ch, _, err := s.reader.ReadRune()
		if err != nil {
			return "", fmt.Errorf("unexpected EOF in quoted string: %w", err)
		}
		if ch == '\'' {
			// Check for doubled single quote (escaped quote)
			nextCh, _, nextErr := s.reader.ReadRune()
			if nextErr == nil && nextCh == '\'' {
				buf.WriteRune('\'')
				continue
			}
			if nextErr == nil {
				s.reader.UnreadRune()
			}
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String(), nil
}

// readWord reads a standard NEXUS word bounded by whitespace or punctuation.
func (s *Scanner) readWord() (string, error) {
	var buf bytes.Buffer
	for {
		ch, _, err := s.reader.ReadRune()
		if err != nil {
			break // EOF reached, which is fine if we have buffered characters
		}
		if unicode.IsSpace(ch) || isPunctuation(ch) || ch == '[' || ch == ']' {
			s.reader.UnreadRune()
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String(), nil
}

// isPunctuation checks if a rune is a standard NEXUS punctuation mark.
func isPunctuation(ch rune) bool {
	switch ch {
	case '(', ')', '{', '}', '[', ']', '/', '\\', ',', ';', ':', '=', '*', '"', '+', '-', '<', '>', '~':
		return true
	default:
		return false
	}
}
