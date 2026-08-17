package idl

import (
	"bytes"
	"reflect"
)

type RawMessage []byte

// Unmarshal defines a custom unmarshal function, as we do not have the type context such as in candid payloads.
func (r RawMessage) Unmarshal(_v any) error {
	v := reflect.ValueOf(_v)
	if v.Kind() != reflect.Pointer {
		return NewUnmarshalGoError(r, _v)
	}
	t, err := TypeOf(v.Elem().Interface())
	if err != nil {
		return err
	}
	vs, err := t.Decode(bytes.NewReader(r), NewBudget(DefaultDecodingQuota))
	if err != nil {
		return err
	}
	if err := UnmarshalGo(t, vs, _v); err != nil {
		return err
	}
	return nil
}
