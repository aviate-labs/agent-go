package certificate

import (
	"crypto/sha256"
	"fmt"
	"unicode/utf8"

	"github.com/fxamacker/cbor/v2"
)

func Serialize(node Node) ([]byte, error) {
	return cbor.Marshal(serialize(node))
}

func domainSeparator(t string) []byte {
	return append(
		[]byte{uint8(len(t))},
		[]byte(t)...,
	)
}

func serialize(node Node) []any {
	switch n := node.(type) {
	case Empty:
		return []any{0x00}
	case Fork:
		return []any{
			0x01,
			serialize(n.LeftTree),
			serialize(n.RightTree),
		}
	case Labeled:
		return []any{
			0x02,
			[]byte(n.Label),
			serialize(n.Tree),
		}
	case Leaf:
		return []any{
			0x03,
			[]byte(n),
		}
	case Pruned:
		return []any{
			0x04,
			n,
		}
	}
	return nil
}

type Empty struct{}

func (e Empty) Reconstruct() [32]byte {
	return sha256.Sum256(domainSeparator("ic-hashtree-empty"))
}

func (e Empty) String() string {
	return "∅"
}

type Fork struct {
	LeftTree  Node
	RightTree Node
}

func (f Fork) Reconstruct() [32]byte {
	l := f.LeftTree.Reconstruct()
	r := f.RightTree.Reconstruct()
	return sha256.Sum256(append(
		domainSeparator("ic-hashtree-fork"),
		append(l[:], r[:]...)...,
	))
}

func (f Fork) String() string {
	return fmt.Sprintf("{%s|%s}", f.LeftTree, f.RightTree)
}

type Label []byte

func (l Label) String() string {
	return string(l)
}

type Labeled struct {
	Label Label
	Tree  Node
}

func (l Labeled) Reconstruct() [32]byte {
	t := l.Tree.Reconstruct()
	return sha256.Sum256(append(
		domainSeparator("ic-hashtree-labeled"),
		append(l.Label, t[:]...)...,
	))
}

func (l Labeled) String() string {
	return fmt.Sprintf("%s:%s", l.Label, l.Tree)
}

type Leaf []byte

func (l Leaf) Reconstruct() [32]byte {
	return sha256.Sum256(append(
		domainSeparator("ic-hashtree-leaf"),
		l...,
	))
}

func (l Leaf) String() string {
	if utf8.Valid(l) {
		return string(l)
	}
	return fmt.Sprintf("0x%x", []byte(l))
}

type Node interface {
	Reconstruct() [32]byte
	fmt.Stringer
}

func Deserialize(data []byte) (Node, error) {
	var s []any
	if err := cbor.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return DeserializeNode(s)
}

func DeserializeNode(s []any) (Node, error) {
	tag, ok := s[0].(uint64)
	if !ok {
		return nil, fmt.Errorf("unknown tag: %v", s[0])
	}
	switch tag {
	case 0:
		if l := len(s); l != 1 {
			return nil, fmt.Errorf("invalid len: %d", l)
		}
		return Empty{}, nil
	case 1:
		if l := len(s); l != 3 {
			return nil, fmt.Errorf("invalid len: %d", l)
		}
		lt, ok := s[1].([]any)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[1])
		}
		l, err := DeserializeNode(lt)
		if err != nil {
			return nil, err
		}
		rt, ok := s[2].([]any)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[2])
		}
		r, err := DeserializeNode(rt)
		if err != nil {
			return nil, err
		}
		return Fork{
			LeftTree:  l,
			RightTree: r,
		}, nil
	case 2:
		if l := len(s); l != 3 {
			return nil, fmt.Errorf("invalid len: %d", l)
		}
		l, ok := s[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[1])
		}
		rt, ok := s[2].([]any)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[2])
		}
		t, err := DeserializeNode(rt)
		if err != nil {
			return nil, err
		}
		return Labeled{
			Label: l,
			Tree:  t,
		}, nil
	case 3:
		if l := len(s); l != 2 {
			return nil, fmt.Errorf("invalid len: %d", l)
		}
		l, ok := s[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[1])
		}
		return Leaf(l), nil
	case 4:
		if l := len(s); l != 2 {
			return nil, fmt.Errorf("invalid len: %d", l)
		}
		p, ok := s[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("unknown value: %v", s[1])
		}
		if l := len(p); l != 32 {
			return nil, fmt.Errorf("invalid hash len: %d", l)
		}
		var p32 [32]byte
		copy(p32[:], p)
		return Pruned(p32), nil
	default:
		return nil, fmt.Errorf("invalid tag: %d", tag)
	}
}

type Pruned [32]byte

func (p Pruned) Reconstruct() [32]byte {
	return p
}

func (p Pruned) String() string {
	return fmt.Sprintf("0x%x", p[:])
}
