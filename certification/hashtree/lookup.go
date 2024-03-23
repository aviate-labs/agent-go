package hashtree

import (
	"bytes"
	"fmt"
)

// LookupResult is the result of a lookup.
type LookupResult struct {
	// Type is the type of the lookup result.
	Type LookupResultType
	// Value is the value of the leaf. Can be nil if the type is not LookupResultFound.
	Value []byte
}

func lookupPath(n Node, path ...Label) LookupResult {
	switch {
	case len(path) == 0:
		switch n := n.(type) {
		case Leaf:
			return LookupResult{
				Type:  LookupResultFound,
				Value: n,
			}
		case nil, Empty:
			return LookupResult{
				Type: LookupResultAbsent,
			}
		case Pruned:
			return LookupResult{
				Type: LookupResultUnknown,
			}
		default:
			// Labeled, Fork
			return LookupResult{
				Type: LookupResultError,
			}
		}
	default:
		switch l := lookupLabel(n, path[0]); l.Type {
		case lookupLabelResultFound:
			return lookupPath(l.Node, path[1:]...)
		case lookupLabelResultUnknown:
			return LookupResult{
				Type: LookupResultUnknown,
			}
		default:
			return LookupResult{
				Type: LookupResultAbsent,
			}
		}
	}
}

// Found returns an error if the lookup result is not found.
func (r LookupResult) Found() error {
	switch r.Type {
	case LookupResultAbsent:
		return fmt.Errorf("not found")
	case LookupResultUnknown:
		return fmt.Errorf("unknown")
	case LookupResultError:
		return fmt.Errorf("error")
	default:
		return nil
	}
}

// LookupResultType is the type of the lookup result.
// It indicates whether the result is guaranteed to be absent, unknown or found.
type LookupResultType int

const (
	// LookupResultAbsent means that the result is guaranteed to be absent.
	LookupResultAbsent LookupResultType = iota
	// LookupResultUnknown means that the result is unknown, some leaves were pruned.
	LookupResultUnknown
	// LookupResultFound means that the result is found.
	LookupResultFound
	// LookupResultError means that the result is an error, the path is not valid in this context.
	LookupResultError
)

// LookupSubTreeResult is the result of a lookup sub-tree.
type LookupSubTreeResult struct {
	// Type is the type of the lookup sub-tree result.
	Type LookupResultType
	// Node is the node that was found. Can be nil if the type is not LookupResultFound.
	Node Node
}

func lookupSubTree(n Node, path ...Label) LookupSubTreeResult {
	switch {
	case len(path) == 0:
		return LookupSubTreeResult{
			Type: LookupResultFound,
			Node: n,
		}
	default:
		switch l := lookupLabel(n, path[0]); l.Type {
		case lookupLabelResultFound:
			return lookupSubTree(l.Node, path[1:]...)
		case lookupLabelResultUnknown:
			return LookupSubTreeResult{
				Type: LookupResultUnknown,
			}
		default:
			return LookupSubTreeResult{
				Type: LookupResultAbsent,
			}
		}
	}
}

// Found returns an error if the lookup sub-tree result is not found.
func (r LookupSubTreeResult) Found() error {
	switch r.Type {
	case LookupResultAbsent:
		return fmt.Errorf("not found")
	case LookupResultUnknown:
		return fmt.Errorf("unknown")
	case LookupResultError:
		return fmt.Errorf("error")
	default:
		return nil
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
