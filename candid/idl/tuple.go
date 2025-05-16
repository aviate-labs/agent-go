package idl

import (
	"fmt"
	"reflect"
	"strings"
)

func isTupleType(value any) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := ParseTags(field)
			if tag.TupleType {
				return true
			}
		}
	}
	return false
}

func NewTupleType(fields map[string]Type) *RecordType {
	var rec RecordType
	for i := range len(fields) {
		id := fmt.Sprintf("%d", i)
		rec.Fields = append(rec.Fields, FieldType{
			Type: fields[id],
		})
	}
	rec.IsTuple = true
	return &rec
}

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
