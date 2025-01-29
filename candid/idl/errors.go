package idl

import "fmt"

type DecodeError struct {
	Types       TupleType
	Description string
}

func (e DecodeError) Error() string {
	return fmt.Sprintf("%s %s", e.Types.String(), e.Description)
}

type EncodeValueError struct {
	Expected OpCode
	Value    any
}

func NewEncodeValueError(v any, o OpCode) *EncodeValueError {
	return &EncodeValueError{
		Expected: o,
		Value:    v,
	}
}

func (e EncodeValueError) Error() string {
	return fmt.Sprintf("invalid type %v (%T), expected type %s", e.Value, e.Value, e.Expected)
}

type FormatError struct {
	Description string
}

func (e FormatError) Error() string {
	return fmt.Sprintf("() %s", e.Description)
}

type UnmarshalGoError struct {
	Raw any
	V   any
}

func NewUnmarshalGoError(raw any, v any) *UnmarshalGoError {
	return &UnmarshalGoError{
		Raw: raw,
		V:   v,
	}
}

func (e UnmarshalGoError) Error() string {
	return fmt.Sprintf("cannot unmarshal %v (%T) into Go value of type %T", e.Raw, e.Raw, e.V)
}
