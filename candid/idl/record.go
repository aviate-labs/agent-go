package idl

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/aviate-labs/leb128"
)

func StructToMap(value any) (map[string]any, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Struct:
		m := make(map[string]any)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := ParseTags(field)
			m[tag.Name] = v.Field(i).Interface()
		}
		return m, nil
	default:
		return nil, fmt.Errorf("invalid value kind: %s", v.Kind())
	}
}

type RecordType struct {
	Fields []FieldType
}

func NewRecordType(fields map[string]Type) *RecordType {
	var rec RecordType
	for k, v := range fields {
		rec.Fields = append(rec.Fields, FieldType{
			Name: k,
			Type: v,
		})
	}
	sort.Slice(rec.Fields, func(i, j int) bool {
		return Hash(rec.Fields[i].Name).Cmp(Hash(rec.Fields[j].Name)) < 0
	})
	return &rec
}

func (record RecordType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	for _, f := range record.Fields {
		if err := f.Type.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}

	id, err := leb128.EncodeSigned(big.NewInt(recType))
	if err != nil {
		return err
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(record.Fields))))
	if err != nil {
		return err
	}
	var vs []byte
	for _, f := range record.Fields {
		l, err := leb128.EncodeUnsigned(Hash(f.Name))
		if err != nil {
			return nil
		}
		t, err := f.Type.EncodeType(tdt)
		if err != nil {
			return nil
		}
		vs = append(vs, concat(l, t)...)
	}

	tdt.Add(record, concat(id, l, vs))
	return nil
}

func (record RecordType) Decode(r_ *bytes.Reader) (any, error) {
	rec := make(map[string]any)
	for _, f := range record.Fields {
		v, err := f.Type.Decode(r_)
		if err != nil {
			return nil, err
		}
		rec[f.Name] = v
	}
	if len(rec) == 0 {
		return nil, nil
	}
	return rec, nil
}

func (record RecordType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[record.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", record)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (record RecordType) EncodeValue(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	fs, ok := v.(map[string]any)
	if !ok {
		var err error
		if fs, err = StructToMap(v); err != nil {
			return nil, NewEncodeValueError(v, recType)
		}
	}
	var vs_ []any
	for _, f := range record.Fields {
		vs_ = append(vs_, fs[f.Name])
	}
	var vs []byte
	for i, f := range record.Fields {
		v_, err := f.Type.EncodeValue(vs_[i])
		if err != nil {
			return nil, err
		}
		vs = append(vs, v_...)
	}
	return vs, nil
}

func (record RecordType) String() string {
	var s []string
	for _, f := range record.Fields {
		s = append(s, fmt.Sprintf("%s:%s", f.Name, f.Type.String()))
	}
	return fmt.Sprintf("record {%s}", strings.Join(s, "; "))
}

func (record RecordType) UnmarshalGo(raw any, _v any) error {
	if raw == nil && record.Fields == nil {
		return nil // Empty record.
	}

	m := make(map[string]any)
	switch rv := reflect.ValueOf(raw); rv.Kind() {
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			m[k.String()] = rv.MapIndex(k).Interface()
		}
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Type().Field(i)
			tag := ParseTags(f)
			m[tag.Name] = rv.Field(i).Interface()
		}
	default:
		return NewUnmarshalGoError(raw, _v)
	}

	if v, ok := _v.(*map[string]any); ok {
		return record.unmarshalMap(m, v)
	}

	v, ok := checkIsPtr(_v)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	if v.Kind() != reflect.Struct {
		return NewUnmarshalGoError(raw, _v)
	}
	return record.unmarshalStruct(m, v)
}

func (record RecordType) unmarshalMap(raw map[string]any, _v *map[string]any) error {
	m := make(map[string]any)
	for _, f := range record.Fields {
		v, err := EmptyOf(f.Type)
		if err != nil {
			return err
		}

		r := reflect.New(reflect.ValueOf(v).Type()).Elem()
		if err := f.Type.UnmarshalGo(raw[f.Name], r.Addr().Interface()); err != nil {
			return err
		}
		m[f.Name] = r.Interface()
	}
	*_v = m
	return nil
}

func (record RecordType) unmarshalStruct(raw map[string]any, _v reflect.Value) error {
	findField := func(name string) (reflect.Value, bool) {
		name = lowerFirstCharacter(name)
		for i := 0; i < _v.NumField(); i++ {
			f := _v.Type().Field(i)
			tag := ParseTags(f)
			if tag.Name == name || HashString(tag.Name) == name {
				return _v.Field(i), true
			}
		}
		return reflect.Value{}, false
	}
	for _, f := range record.Fields {
		v, ok := findField(f.Name)
		if !ok {
			return NewUnmarshalGoError(raw, _v.Interface())
		}
		v = v.Addr()
		if v.IsNil() {
			// Set to a new value if the field is nil.
			if _, ok := f.Type.(*NullType); !ok {
				v.Set(reflect.New(v.Type().Elem()))
			}
		}

		if err := f.Type.UnmarshalGo(raw[f.Name], v.Interface()); err != nil {
			return err
		}
	}
	return nil
}
