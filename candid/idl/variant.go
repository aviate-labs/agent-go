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

func isVariantType(value any) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := ParseTags(field)
			if tag.VariantType {
				return true
			}
		}
	}
	return false
}

type Variant struct {
	Name  string
	Value any
	Type  Type
}

type VariantType struct {
	Fields []FieldType
}

func NewVariantType(fields map[string]Type) *VariantType {
	var variant VariantType
	for k, v := range fields {
		variant.Fields = append(variant.Fields, FieldType{
			Name: k,
			Type: v,
		})
	}
	sort.Slice(variant.Fields, func(i, j int) bool {
		return Hash(variant.Fields[i].Name).Cmp(Hash(variant.Fields[j].Name)) < 0
	})
	return &variant
}

func (variant VariantType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	for _, f := range variant.Fields {
		if err := f.Type.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}

	id, err := leb128.EncodeSigned(VarOpCode.BigInt())
	if err != nil {
		return err
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(variant.Fields))))
	if err != nil {
		return err
	}
	var vs []byte
	for _, f := range variant.Fields {
		id, err := leb128.EncodeUnsigned(Hash(f.Name))
		if err != nil {
			return nil
		}
		t, err := f.Type.EncodeType(tdt)
		if err != nil {
			return nil
		}
		vs = append(vs, concat(id, t)...)
	}

	tdt.Add(variant, concat(id, l, vs))
	return nil
}

func (variant VariantType) Decode(r *bytes.Reader) (any, error) {
	id, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	if id.Cmp(big.NewInt(int64(len(variant.Fields)))) >= 0 {
		return nil, fmt.Errorf("invalid variant index: %v", id)
	}
	f := variant.Fields[int(id.Int64())]
	v_, err := f.Type.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Variant{
		Name:  f.Name,
		Value: v_,
		Type:  f.Type,
	}, nil
}

func (variant VariantType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[variant.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", variant)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (variant VariantType) EncodeValue(value any) ([]byte, error) {
	fs, ok := value.(Variant)
	if !ok {
		v, err := variant.structToVariant(value)
		if err != nil {
			return nil, err
		}
		return variant.EncodeValue(*v)
	}
	for i, f := range variant.Fields {
		if f.Name == fs.Name {
			id, err := leb128.EncodeUnsigned(big.NewInt(int64(i)))
			if err != nil {
				return nil, err
			}
			v_, err := f.Type.EncodeValue(fs.Value)
			if err != nil {
				return nil, err
			}
			return concat(id, v_), nil
		}
	}
	return nil, fmt.Errorf("unknown variant: %v", value)
}

func (variant VariantType) String() string {
	var s []string
	for _, f := range variant.Fields {
		s = append(s, fmt.Sprintf("%s:%s", f.Name, f.Type.String()))
	}
	return fmt.Sprintf("variant {%s}", strings.Join(s, "; "))
}

func (variant VariantType) UnmarshalGo(raw any, _v any) error {
	var (
		name  string
		value any
	)
	switch raw := raw.(type) {
	case *Variant:
		name = raw.Name
		value = raw.Value
	default:
		switch rv := reflect.ValueOf(raw); rv.Kind() {
		case reflect.Map:
			if rv.Len() != 1 {
				return NewUnmarshalGoError(raw, _v)
			}
			for _, k := range rv.MapKeys() {
				name = k.String()
				value = rv.MapIndex(k).Interface()
			}
		case reflect.Struct:
			if rv.NumField() != 1 {
				return NewUnmarshalGoError(raw, _v)
			}
			for i := 0; i < rv.NumField(); i++ {
				f := rv.Type().Field(i)
				tag := ParseTags(f)
				v := rv.Field(i)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				name = tag.Name
				value = v.Interface()
			}
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	}

	if v, ok := _v.(*map[string]any); ok {
		return variant.unmarshalMap(name, value, v)
	}

	v, ok := checkIsPtr(_v)
	if !ok {
		return NewUnmarshalGoError(raw, _v)
	}
	if v.Kind() != reflect.Struct {
		return NewUnmarshalGoError(raw, _v)
	}
	return variant.unmarshalStruct(name, value, v)
}

func (variant VariantType) structToVariant(value any) (*Variant, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := ParseTags(field)
			if !tag.VariantType {
				return nil, fmt.Errorf("invalid variant field: %s", v.Type())
			}
			if !v.Field(i).IsNil() {
				return &Variant{
					Name:  tag.Name,
					Value: v.Field(i).Elem().Interface(),
				}, nil
			}
			if i == v.NumField()-1 {
				return nil, fmt.Errorf("invalid variant: no variant selected")
			}
		}
		return nil, fmt.Errorf("invalid variant: %s", v.Type())
	default:
		return nil, fmt.Errorf("invalid value kind: %s", v.Kind())
	}
}

func (variant VariantType) unmarshalMap(name string, value any, _v *map[string]any) error {
	for _, f := range variant.Fields {
		if f.Name != name {
			continue
		}

		v, err := EmptyOf(f.Type)
		if err != nil {
			return NewUnmarshalGoError(value, _v)
		}

		r := reflect.New(reflect.ValueOf(v).Type()).Elem()
		if err := f.Type.UnmarshalGo(value, r.Addr().Interface()); err != nil {
			return err
		}
		*_v = map[string]any{
			f.Name: r.Interface(),
		}
		return nil
	}
	return NewUnmarshalGoError(value, _v)
}

func (variant VariantType) unmarshalStruct(name string, value any, _v reflect.Value) error {
	var v reflect.Value
	name = lowerFirstCharacter(name)
	for i := 0; i < _v.NumField(); i++ {
		tag := ParseTags(_v.Type().Field(i))
		if tag.Name == name || HashString(tag.Name) == name {
			v = _v.Field(i)
		}
	}
	if !v.IsValid() {
		return NewUnmarshalGoError(value, _v.Interface())
	}
	for _, f := range variant.Fields {
		if f.Name != name {
			continue
		}

		if v.Kind() != reflect.Ptr {
			return NewUnmarshalGoError(value, _v.Interface())
		}
		if v.IsNil() {
			// Allocate a new value if the pointer is nil.
			v.Set(reflect.New(v.Type().Elem()))
		}
		return f.Type.UnmarshalGo(value, v.Interface())
	}
	return NewUnmarshalGoError(value, _v.Interface())
}
