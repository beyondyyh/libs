package conhash

// Hash node struct
type Node struct {
	ident    string
	replicas int
}

// Create new node
func NewNode(addr string, replicas int) *Node {
	return &Node{
		ident:    addr,
		replicas: replicas,
	}
}

// Get node ident string
func (n *Node) GetIdent() string {
	return n.ident
}

// String return node ident
func (n *Node) String() string {
	return n.ident
}

// Get node replicas
func (n *Node) GetReplicas() int {
	return n.replicas
}
