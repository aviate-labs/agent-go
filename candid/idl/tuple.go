package idl

import (
	"fmt"
	"strings"
)

// TupleType is a collection of types.
type TupleType []Type

// String returns the string representation of the type.
func (ts TupleType) String() string {
	var s []string
	for _, t := range ts {
		s = append(s, t.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(s, ", "))
}
