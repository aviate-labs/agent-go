package hashtree

import "fmt"

// Lookup looks up a path in the node.
func Lookup(n Node, path ...Label) ([]byte, error) {
	return lookupPath(n, path, 0)
}

// HashTree is a hash tree.
type HashTree struct {
	Root Node
}

// NewHashTree creates a new hash tree.
func NewHashTree(root Node) HashTree {
	return HashTree{root}
}

// Digest returns the digest of the hash tree.
func (t HashTree) Digest() [32]byte {
	return t.Root.Reconstruct()
}

// Lookup looks up a path in the hash tree.
func (t HashTree) Lookup(path ...Label) ([]byte, error) {
	return Lookup(t.Root, path...)
}

// LookupSubTree looks up a path in the hash tree and returns the sub-tree.
func (t HashTree) LookupSubTree(path ...Label) (Node, error) {
	return LookupSubTree(t.Root, path...)
}

// MarshalCBOR marshals a hash tree.
func (t HashTree) MarshalCBOR() ([]byte, error) {
	return Serialize(t.Root)
}

// UnmarshalCBOR unmarshals a hash tree.
func (t *HashTree) UnmarshalCBOR(bytes []byte) error {
	root, err := Deserialize(bytes)
	if err != nil {
		return err
	}
	t.Root = root
	return nil
}

// LookupSubTree looks up a path in the node and returns the sub-tree.
func LookupSubTree(n Node, path ...Label) (Node, error) {
	return lookupSubTree(n, path, 0)
}

type PathValuePair[V any] struct {
	Path  []Label
	Value V
}

func AllChildren(n Node) ([]PathValuePair[Node], error) {
	return allChildren(n)
}

// AllPaths returns all non-empty labeled paths in the hash tree, does not include pruned nodes.
func AllPaths(n Node) ([]PathValuePair[[]byte], error) {
	return allLabeled(n, nil)
}

func allChildren(n Node) ([]PathValuePair[Node], error) {
	switch n := n.(type) {
	case Empty, Pruned, Leaf:
		return nil, nil
	case Labeled:
		return []PathValuePair[Node]{{Path: []Label{n.Label}, Value: n.Tree}}, nil
	case Fork:
		left, err := allChildren(n.LeftTree)
		if err != nil {
			return nil, err
		}
		right, err := allChildren(n.RightTree)
		if err != nil {
			return nil, err
		}
		return append(left, right...), nil
	default:
		return nil, fmt.Errorf("unsupported node type: %T", n)
	}
}

func allLabeled(n Node, path []Label) ([]PathValuePair[[]byte], error) {
	switch n := n.(type) {
	case Empty, Pruned:
		return nil, nil
	case Leaf:
		return []PathValuePair[[]byte]{{Path: path, Value: n}}, nil
	case Labeled:
		return allLabeled(n.Tree, append(path, n.Label))
	case Fork:
		left, err := allLabeled(n.LeftTree, path)
		if err != nil {
			return nil, err
		}
		right, err := allLabeled(n.RightTree, path)
		if err != nil {
			return nil, err
		}
		return append(left, right...), nil
	default:
		return nil, fmt.Errorf("unsupported node type: %T", n)
	}
}
