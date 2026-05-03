package trees

import (
	"strings"

	"github.com/espinosajuanma/nexus/parser"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse implements the Block interface for TreesBlock.
func (t *TreesBlock) Parse(s *scanner.Scanner) error {
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
		case "TRANSLATE":
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			t.parseTranslate(tokens)
		case "TREE":
			if err := t.parseTree(s); err != nil {
				return err
			}
		default:
			if _, err := parser.ReadUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

func (t *TreesBlock) parseTranslate(tokens []string) {
	var currentToken string
	for _, tok := range tokens {
		if tok == "," {
			continue
		}
		if currentToken == "" {
			currentToken = tok
		} else {
			t.Translate[currentToken] = tok
			currentToken = ""
		}
	}
}

func (t *TreesBlock) parseTree(s *scanner.Scanner) error {
	var tree Tree

	nextToken, err := s.NextToken()
	if err != nil {
		return err
	}

	if nextToken == "*" {
		tree.IsDefault = true
		nextToken, err = s.NextToken()
		if err != nil {
			return err
		}
	}

	tree.Name = nextToken

	// Skip the '=' token
	if _, err := s.NextToken(); err != nil {
		return err
	}

	// Read the Newick tokens until ';'
	specTokens, err := parser.ReadUntilSemicolon(s)
	if err != nil {
		return err
	}

	// Build the actual AST
	tree.Root = buildNewickTree(specTokens)
	t.Trees = append(t.Trees, tree)

	return nil
}

// buildNewickTree uses a stack to convert a flat slice of tokens into a nested node structure.
func buildNewickTree(tokens []string) *TreeNode {
	root := &TreeNode{}
	current := root
	var stack []*TreeNode

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		// Catch NEXUS comments like [&R] or [&U]
		if strings.HasPrefix(tok, "[") {
			current.Comments = append(current.Comments, tok)
			continue
		}

		switch tok {
		case "(":
			// Create a new child, add it to current, and push current to stack
			child := &TreeNode{}
			current.Children = append(current.Children, child)
			stack = append(stack, current)
			current = child
		case ",":
			// Create a sibling. The parent is on top of the stack.
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				child := &TreeNode{}
				parent.Children = append(parent.Children, child)
				current = child
			}
		case ")":
			// End of current clade. Step back up to the parent.
			if len(stack) > 0 {
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1] // Pop the stack
			}
		case ":":
			// Branch length follows the colon
			if i+1 < len(tokens) {
				current.BranchLength = tokens[i+1]
				i++ // Skip the next token since we consumed it
			}
		default:
			current.Name = tok
		}
	}

	return root
}
