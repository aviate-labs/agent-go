package hashtree

import (
	"bytes"
)

func lookupPath(n Node, path []Label, idx int) ([]byte, error) {
	switch {
	case len(path) == 0 || len(path) == idx:
		switch n := n.(type) {
		case Leaf:
			return n, nil
		case nil, Empty:
			return nil, NewLookupAbsentError(path, idx-1)
		case Pruned:
			return nil, NewLookupUnknownError(path, idx-1)
		default:
			// Labeled, Fork
			return nil, NewLookupError(path, idx-1)
		}
	default:
		switch l := lookupLabel(n, path[idx]); l.Type {
		case lookupLabelResultFound:
			return lookupPath(l.Node, path, idx+1)
		case lookupLabelResultUnknown:
			return nil, NewLookupUnknownError(path, idx)
		default:
			return nil, NewLookupAbsentError(path, idx)
		}
	}
}

func lookupSubTree(n Node, path []Label, idx int) (Node, error) {
	switch {
	case len(path) == 0 || len(path) == idx:
		return n, nil
	default:
		switch l := lookupLabel(n, path[idx]); l.Type {
		case lookupLabelResultFound:
			return lookupSubTree(l.Node, path, idx+1)
		case lookupLabelResultUnknown:
			return nil, NewLookupUnknownError(path, idx)
		default:
			return nil, NewLookupAbsentError(path, idx)
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
