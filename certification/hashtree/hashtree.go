package hashtree

import "fmt"

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
	return lookupPath(t.Root, path, 0)
}

// LookupSubTree looks up a path in the hash tree and returns the sub-tree.
func (t HashTree) LookupSubTree(path ...Label) (Node, error) {
	return lookupSubTree(t.Root, path, 0)
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

type PathValuePair struct {
	Path  []Label
	Value []byte
}

func AllChildren(n Node) ([]PathValuePair, error) {
	return allChildren(n)
}

// AllPaths returns all non-empty labeled paths in the hash tree, does not include pruned nodes.
func AllPaths(n Node) ([]PathValuePair, error) {
	return allLabeled(n, nil)
}

func allChildren(n Node) ([]PathValuePair, error) {
	switch n := n.(type) {
	case Empty, Pruned, Leaf:
		return nil, nil
	case Labeled:
		switch c := n.Tree.(type) {
		case Leaf:
			return []PathValuePair{{Path: []Label{n.Label}, Value: c}}, nil
		default:
			return nil, nil
		}
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

func allLabeled(n Node, path []Label) ([]PathValuePair, error) {
	switch n := n.(type) {
	case Empty, Pruned:
		return nil, nil
	case Leaf:
		return []PathValuePair{{Path: path, Value: n}}, nil
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
