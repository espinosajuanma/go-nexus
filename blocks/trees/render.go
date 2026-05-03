package trees

import (
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/espinosajuanma/nexus/core"
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

	var buf bytes.Buffer
	if err := treesTmpl.Execute(&buf, templateData); err != nil {
		return "[ERROR rendering TREES block: " + err.Error() + "]\n"
	}
	return buf.String()
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
