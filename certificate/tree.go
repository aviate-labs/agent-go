package certificate

type HashTree struct {
	root Node
}

func NewHashTree(root Node) HashTree {
	return HashTree{root}
}

func (t HashTree) Digest() [32]byte {
	return t.root.Reconstruct()
}
