package hashtree

import (
	"fmt"
	"strings"
)

// pathToString converts a path to a string, by joining the (string) labels with a slash.
func pathToString(path []Label) string {
	var sb strings.Builder
	for i, p := range path {
		if i > 0 {
			sb.WriteByte('/')
		}
		sb.WriteString(string(p))
	}
	return sb.String()
}

// LookupError is an error that occurs during a lookup.
type LookupError struct {
	// Type is the type of the lookup result.
	Type LookupResultType
	// Path is the path that was looked up.
	Path []Label
	// Index is the index in the path where the error occurred.
	Index int
}

// NewLookupAbsentError returns a new LookupError with type LookupResultAbsent.
func NewLookupAbsentError(path []Label, index int) LookupError {
	return LookupError{
		Type:  LookupResultAbsent,
		Path:  path,
		Index: index,
	}
}

// NewLookupError returns a new LookupError with type LookupResultError.
func NewLookupError(path []Label, index int) LookupError {
	return LookupError{
		Type:  LookupResultError,
		Path:  path,
		Index: index,
	}
}

// NewLookupUnknownError returns a new LookupError with type LookupResultUnknown.
func NewLookupUnknownError(path []Label, index int) LookupError {
	return LookupError{
		Type:  LookupResultUnknown,
		Path:  path,
		Index: index,
	}
}

func (l LookupError) Error() string {
	return fmt.Sprintf("lookup error (path: %q) at %q: %s", pathToString(l.Path), l.Path[l.Index], l.error())
}

func (l LookupError) error() string {
	switch l.Type {
	case LookupResultAbsent:
		return "not found, not present in the tree"
	case LookupResultUnknown:
		return "not found, could be pruned"
	case LookupResultError:
		return "error, can not exist in the tree"
	default:
		return "unknown lookup error"
	}
}

// LookupResultType is the type of the lookup result.
// It indicates whether the result is guaranteed to be absent, unknown or is an invalid tree.
type LookupResultType int

const (
	// LookupResultAbsent means that the result is guaranteed to be absent.
	LookupResultAbsent LookupResultType = iota
	// LookupResultUnknown means that the result is unknown, some leaves were pruned.
	LookupResultUnknown
	// LookupResultError means that the result is an error, the path is not valid in this context.
	LookupResultError
)
