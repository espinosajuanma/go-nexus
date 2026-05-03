package trees

import (
	_ "embed"
	"sort"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/templater"
)

//go:embed trees.tmpl
var treesTmpl string

// Render implements the Block interface for TreesBlock.
func (t *TreesBlock) Render() (string, error) {
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
			Taxon:  core.EncodeName(t.Translate[tok]),
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

	tmpl, err := templater.New("trees", treesTmpl)
	if err != nil {
		return "", err
	}
	rendered, err := tmpl.Render(templateData)

	return rendered, err
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
