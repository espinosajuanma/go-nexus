package trees

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

// NewNode creates a fresh TreeNode.
func NewNode() *TreeNode {
	return &TreeNode{}
}

// SetName assigns a label to the node (a taxon name or a clade name)
// Returns the node to allow method chaining.
func (n *TreeNode) SetName(name string) *TreeNode {
	n.Name = name
	return n
}

// SetBranchLength assigns the branch length below this node
// Returns the node to allow method chaining.
func (n *TreeNode) SetBranchLength(length string) *TreeNode {
	n.BranchLength = length
	return n
}

// AddComment appends a NEXUS comment to the node (e.g., "[&U]" for unrooted)
// Returns the node to allow method chaining.
func (n *TreeNode) AddComment(comment string) *TreeNode {
	n.Comments = append(n.Comments, comment)
	return n
}

// AddChild attaches a sub-clade or leaf node to this node
// Returns the PARENT node to allow method chaining.
func (n *TreeNode) AddChild(child *TreeNode) *TreeNode {
	n.Children = append(n.Children, child)
	return n
}
