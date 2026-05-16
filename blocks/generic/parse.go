package generic

import (
	"fmt"
	"strings"

	"github.com/espinosajuanma/go-nexus/scanner"
)

// Parse implements the Block interface for GenericBlock.
func (t *GenericBlock) Parse(s *scanner.Scanner) error {
	content, err := s.ReadRawUntilBlockEnd()
	if err != nil {
		return fmt.Errorf("error reading content for GENERIC block: %w", err)
	}

	// Clean up surrounding whitespaces so it sits perfectly inside the template block
	content = strings.TrimRight(content, " \t\r\n")
	content = strings.TrimPrefix(content, "\r\n")
	content = strings.TrimPrefix(content, "\n")

	t.Content = content
	return nil
}
