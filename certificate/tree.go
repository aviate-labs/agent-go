package certificate

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
