package certificate

import "bytes"

func Lookup(path [][]byte, node Node) []byte {
	if len(path) == 0 {
		switch n := node.(type) {
		case Leaf:
			return n
		default:
			return nil
		}
	}

	n := findLabel(flattenNode(node), path[0])
	if n != nil {
		return Lookup(path[1:], *n)
	}
	return nil
}

type LabelResult string

const (
	Absent   LabelResult = "absent"
	Continue LabelResult = "continue"
	Found    LabelResult = "found"
	Unknown  LabelResult = "unknown"
)

func findLabel(nodes []Node, label Label) *Node {
	for _, node := range nodes {
		switch n := node.(type) {
		case Labeled:
			if bytes.Equal(label, n.Label) {
				return &n.Tree
			}
		}
	}
	return nil
}

func flattenNode(node Node) []Node {
	switch n := node.(type) {
	case Empty:
		return nil
	case Fork:
		return append(
			flattenNode(n.LeftTree),
			flattenNode(n.RightTree)...,
		)
	default:
		return []Node{node}
	}
}
