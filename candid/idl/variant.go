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

	id, err := leb128.EncodeSigned(big.NewInt(varType))
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
		return nil, fmt.Errorf("invalid variant index: %variant", id)
	}
	v_, err := variant.Fields[int(id.Int64())].Type.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Variant{
		Name:  id.String(),
		Value: v_,
		Type:  variant,
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
		return nil, NewEncodeValueError(variant, varType)
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
	return nil, fmt.Errorf("unknown variant: %variant", value)
}

func (variant VariantType) String() string {
	var s []string
	for _, f := range variant.Fields {
		s = append(s, fmt.Sprintf("%s:%s", f.Name, f.Type.String()))
	}
	return fmt.Sprintf("variant {%s}", strings.Join(s, "; "))
}

func (variant VariantType) UnmarshalGo(raw any, _v any) error {
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
			v := rv.Field(i)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			m[tag.Name] = v.Interface()
		}
	default:
		return NewUnmarshalGoError(raw, _v)
	}
	if len(m) != 1 {
		return NewUnmarshalGoError(raw, _v)
	}

	if v, ok := _v.(*map[string]any); ok {
		return variant.unmarshalMap(m, v)
	}
	v := reflect.ValueOf(_v)
	if v.Kind() != reflect.Ptr {
		return NewUnmarshalGoError(raw, _v)
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return NewUnmarshalGoError(raw, _v)
	}
	return variant.unmarshalStruct(m, v)
}

func (variant VariantType) unmarshalMap(raw map[string]any, _v *map[string]any) error {
	for _, f := range variant.Fields {
		v, err := EmptyOf(f.Type)
		if err != nil {
			continue
		}

		r := reflect.New(reflect.ValueOf(v).Type()).Elem()
		if err := f.Type.UnmarshalGo(raw[f.Name], r.Addr().Interface()); err != nil {
			return err
		}
		*_v = map[string]any{
			f.Name: r.Interface(),
		}
		return nil
	}
	return NewUnmarshalGoError(raw, _v)
}

func (variant VariantType) unmarshalStruct(raw map[string]any, _v reflect.Value) error {
	findField := func(name string) (reflect.Value, bool) {
		name = lowerFirstCharacter(name)
		for i := 0; i < _v.NumField(); i++ {
			f := _v.Type().Field(i)
			tag := ParseTags(f)
			if tag.Name == name {
				return _v.Field(i), true
			}
		}
		return reflect.Value{}, false
	}
	for _, f := range variant.Fields {
		v, ok := findField(f.Name)
		if !ok {
			continue
		}
		if v.Kind() != reflect.Ptr {
			return NewUnmarshalGoError(raw, _v)
		}
		if v.IsNil() {
			// Allocate a new value if the pointer is nil.
			v.Set(reflect.New(v.Type().Elem()))
		}
		if err := f.Type.UnmarshalGo(raw[f.Name], v.Interface()); err != nil {
			return err
		}
		return nil
	}
	return NewUnmarshalGoError(raw, _v)
}
