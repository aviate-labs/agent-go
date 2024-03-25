package hashtree

import (
	"bytes"
)

func lookupPath(n Node, path ...Label) ([]byte, error) {
	switch {
	case len(path) == 0:
		switch n := n.(type) {
		case Leaf:
			return n, nil
		case nil, Empty:
			return nil, NewLookupAbsentError()
		case Pruned:
			return nil, NewLookupUnknownError()
		default:
			// Labeled, Fork
			return nil, NewLookupError()
		}
	default:
		switch l := lookupLabel(n, path[0]); l.Type {
		case lookupLabelResultFound:
			return lookupPath(l.Node, path[1:]...)
		case lookupLabelResultUnknown:
			return nil, NewLookupUnknownError(path...)
		default:
			return nil, NewLookupAbsentError(path...)
		}
	}
}

func lookupSubTree(n Node, path ...Label) (Node, error) {
	switch {
	case len(path) == 0:
		return n, nil
	default:
		switch l := lookupLabel(n, path[0]); l.Type {
		case lookupLabelResultFound:
			return lookupSubTree(l.Node, path[1:]...)
		case lookupLabelResultUnknown:
			return nil, NewLookupUnknownError(path...)
		default:
			return nil, NewLookupAbsentError(path...)
		}
	}
}

// lookupLabelResult is the result of a lookup label.
type lookupLabelResult struct {
	// Type is the type of the lookup label result.
	Type lookupLabelResultType
	// Node is the node that was found. Can be nil.
	Node Node
}

func lookupLabel(n Node, label Label) lookupLabelResult {
	switch n := n.(type) {
	case Labeled:
		c := bytes.Compare(label, n.Label)
		switch {
		case c < 0:
			return lookupLabelResult{
				Type: lookupLabelResultLess,
			}
		case c > 0:
			return lookupLabelResult{
				Type: lookupLabelResultGreater,
			}
		default:
			return lookupLabelResult{
				Type: lookupLabelResultFound,
				Node: n.Tree,
			}
		}
	case Pruned:
		return lookupLabelResult{
			Type: lookupLabelResultUnknown,
		}
	case Fork:
		switch ll := lookupLabel(n.LeftTree, label); ll.Type {
		case lookupLabelResultGreater:
			// Continue looking in the right tree.
			switch rl := lookupLabel(n.RightTree, label); rl.Type {
			case lookupLabelResultLess:
				return lookupLabelResult{
					Type: lookupLabelResultAbsent,
				}
			default:
				return rl
			}
		case lookupLabelResultUnknown:
			// Continue looking in the right tree.
			switch rl := lookupLabel(n.RightTree, label); rl.Type {
			case lookupLabelResultLess:
				return lookupLabelResult{
					Type: lookupLabelResultUnknown,
				}
			default:
				return rl
			}
		default:
			return ll
		}
	default:
		return lookupLabelResult{
			Type: lookupLabelResultAbsent,
		}
	}
}

// lookupLabelResultType is the type of the lookup label result.
// It indicates whether the label is guaranteed to be absent, unknown, less, greater or found.
type lookupLabelResultType int

const (
	// lookupLabelResultAbsent means that the label is absent.
	lookupLabelResultAbsent lookupLabelResultType = iota
	// lookupLabelResultUnknown means that the label is unknown, some leaves were pruned.
	lookupLabelResultUnknown
	// lookupLabelResultLess means that the label was not found, but could be on the left side.
	lookupLabelResultLess
	// lookupLabelResultGreater means that the label was not found, but could be on the right side.
	lookupLabelResultGreater
	// lookupLabelResultFound means that the label was found.
	lookupLabelResultFound
)
