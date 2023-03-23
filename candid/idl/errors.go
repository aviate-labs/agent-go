package idl

import (
	"fmt"
)

type DecodeError struct {
	Types       TupleType
	Description string
}

func (e DecodeError) Error() string {
	return fmt.Sprintf("%s %s", e.Types.String(), e.Description)
}

type EncodeValueError struct {
	Expected int64
	Value    any
}

func NewEncodeValueError(v any, e int64) *EncodeValueError {
	return &EncodeValueError{
		Expected: e,
		Value:    v,
	}
}

func (e EncodeValueError) Error() string {
	return fmt.Sprintf("invalid type %v (%T), expected type %s", e.Value, e.Value, idlString(e.Expected))
}

type FormatError struct {
	Description string
}

func (e FormatError) Error() string {
	return fmt.Sprintf("() %s", e.Description)
}
