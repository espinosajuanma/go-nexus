package nexus

import (
	"bytes"
	"sort"
	"strings"
	"text/template"
)

// The template for the TREES block
const treesTmplStr = `BEGIN TREES;
{{- if .SortedTranslate}}
	TRANSLATE
{{- range $index, $element := .SortedTranslate}}
		{{$element.Token}} {{$element.Taxon}}{{if not $element.IsLast}},{{end}}
{{- end}}
	;
{{- end}}
{{- range .Trees}}
	TREE {{if .IsDefault}}* {{end}}{{.Name}} = {{.Root.Render}};
{{- end}}
END;
`

var treesTmpl = template.Must(template.New("trees").Parse(treesTmplStr))

func init() {
	RegisterBlock("TREES", func() Block {
		return &TreesBlock{
			Translate: make(map[string]string),
		}
	})
}

// TreesBlock stores information about trees.
type TreesBlock struct {
	nexus     *Nexus
	Translate map[string]string
	Trees     []Tree
}

// NewTreesBlock creates, appends, and returns a new TREES block.
func (n *Nexus) NewTreesBlock() *TreesBlock {
	tb := &TreesBlock{
		nexus:     n,
		Translate: make(map[string]string),
	}
	n.Blocks = append(n.Blocks, tb)
	return tb
}

// SetNexus implements the NexusAware interface.
func (t *TreesBlock) SetNexus(n *Nexus) {
	t.nexus = n
}

// Tree represents a single parsed phylogenetic tree.
type Tree struct {
	Name      string
	IsDefault bool
	Root      *TreeNode // The top-level node of the tree
}

// TreeNode represents a single clade or leaf in the tree.
type TreeNode struct {
	Name         string
	BranchLength string
	Comments     []string    // E.g., [&U] or [&R]
	Children     []*TreeNode // Nested sub-clades
}

// Render recursively rebuilds the Newick string from the node structure.
func (n *TreeNode) Render() string {
	if n == nil {
		return ""
	}
	var b strings.Builder

	// Render comments attached to this node (usually [&U] or [&R] at the root)
	for _, c := range n.Comments {
		b.WriteString(c + " ")
	}

	// Recursively render children
	if len(n.Children) > 0 {
		b.WriteString("(")
		for i, child := range n.Children {
			b.WriteString(child.Render())
			if i < len(n.Children)-1 {
				b.WriteString(",")
			}
		}
		b.WriteString(")")
	}

	b.WriteString(n.Name)

	if n.BranchLength != "" {
		b.WriteString(":" + n.BranchLength)
	}

	return strings.TrimSpace(b.String())
}

// Parse implements the Block interface for TreesBlock.
func (t *TreesBlock) Parse(s *Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		if cmd == "END" || cmd == "ENDBLOCK" {
			return expectSemicolon(s)
		}

		switch cmd {
		case "TRANSLATE":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			t.parseTranslate(tokens)
		case "TREE":
			if err := t.parseTree(s); err != nil {
				return err
			}
		default:
			if _, err := readUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// Render implements the Block interface for TreesBlock.
func (t *TreesBlock) Render() string {
	type translatePair struct {
		Token  string
		Taxon  string
		IsLast bool
	}

	var sortedTranslate []translatePair
	var tokens []string
	for k := range t.Translate {
		tokens = append(tokens, k)
	}
	sort.Strings(tokens)

	for i, tok := range tokens {
		sortedTranslate = append(sortedTranslate, translatePair{
			Token:  tok,
			Taxon:  t.Translate[tok],
			IsLast: i == len(tokens)-1,
		})
	}

	templateData := struct {
		SortedTranslate []translatePair
		Trees           []Tree
	}{
		SortedTranslate: sortedTranslate,
		Trees:           t.Trees,
	}

	var buf bytes.Buffer
	if err := treesTmpl.Execute(&buf, templateData); err != nil {
		return "[ERROR rendering TREES block: " + err.Error() + "]\n"
	}
	return buf.String()
}

// AddTranslate maps an arbitrary token (like "1") to a valid taxon name .
// It automatically syncs with the TAXA block.
func (t *TreesBlock) AddTranslate(token string, taxonName string) {
	if t.Translate == nil {
		t.Translate = make(map[string]string)
	}
	t.Translate[token] = taxonName

	// Auto-register the taxon in the TAXA block
	if t.nexus != nil {
		t.nexus.RegisterTaxon(taxonName)
	}
}

// AddTree appends a fully built Tree to the block.
func (t *TreesBlock) AddTree(name string, isDefault bool, root *TreeNode) {
	t.Trees = append(t.Trees, Tree{
		Name:      name,
		IsDefault: isDefault,
		Root:      root,
	})
}

// NewNode creates a fresh TreeNode.
func NewNode() *TreeNode {
	return &TreeNode{}
}

// SetName assigns a label to the node (a taxon name or a clade name) [cite: 1089-1090].
// Returns the node to allow method chaining.
func (n *TreeNode) SetName(name string) *TreeNode {
	n.Name = name
	return n
}

// SetBranchLength assigns the branch length below this node[cite: 1097].
// Returns the node to allow method chaining.
func (n *TreeNode) SetBranchLength(length string) *TreeNode {
	n.BranchLength = length
	return n
}

// AddComment appends a NEXUS comment to the node (e.g., "[&U]" for unrooted) [cite: 1098-1102].
// Returns the node to allow method chaining.
func (n *TreeNode) AddComment(comment string) *TreeNode {
	n.Comments = append(n.Comments, comment)
	return n
}

// AddChild attaches a sub-clade or leaf node to this node [cite: 1081-1082].
// Returns the PARENT node to allow method chaining.
func (n *TreeNode) AddChild(child *TreeNode) *TreeNode {
	n.Children = append(n.Children, child)
	return n
}

// -----------------------------------------------------------------------------
// Parsing Helpers
// -----------------------------------------------------------------------------

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

func (t *TreesBlock) parseTree(s *Scanner) error {
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
	specTokens, err := readUntilSemicolon(s)
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
