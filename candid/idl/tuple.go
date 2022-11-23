package idl

import (
	"fmt"
	"strings"
)

type TupleType []Type

func (ts TupleType) String() string {
	var s []string
	for _, t := range ts {
		s = append(s, t.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(s, ", "))
}
