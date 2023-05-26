package marshal

import (
	"fmt"
	"reflect"
)

type ErrInvalidTypeMatch struct {
	Kind  reflect.Kind
	Value any
}

func NewErrInvalidTypeMatch(v reflect.Value, value any) ErrInvalidTypeMatch {
	return ErrInvalidTypeMatch{
		Kind:  v.Kind(),
		Value: value,
	}
}

func (e ErrInvalidTypeMatch) Error() string {
	return fmt.Sprintf("invalid type match: %q %s", e.Kind, e.Value)
}
