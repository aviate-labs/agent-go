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

func VariantToStruct(r *VariantType, variant *Variant, value any) error {
	v := reflect.ValueOf(value).Elem()
	if !v.CanAddr() {
		return fmt.Errorf("can not address struct value")
	}

	fieldNameToIndex := make(map[string]int)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := parseTags(field)
		fieldNameToIndex[HashString(tag.name)] = i
	}

	i := fieldNameToIndex[r.Fields[0].Name]
	ptrValue := reflect.New(reflect.TypeOf(variant.Value))
	ptrValue.Elem().Set(reflect.ValueOf(variant.Value))
	v.Field(i).Set(ptrValue)
	return nil
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

func (v VariantType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	for _, f := range v.Fields {
		if err := f.Type.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}

	id, err := leb128.EncodeSigned(big.NewInt(varType))
	if err != nil {
		return err
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(v.Fields))))
	if err != nil {
		return err
	}
	var vs []byte
	for _, f := range v.Fields {
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

	tdt.Add(v, concat(id, l, vs))
	return nil
}

func (v VariantType) Decode(r *bytes.Reader) (any, error) {
	id, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	if id.Cmp(big.NewInt(int64(len(v.Fields)))) >= 0 {
		return nil, fmt.Errorf("invalid variant index: %v", id)
	}
	v_, err := v.Fields[int(id.Int64())].Type.Decode(r)
	if err != nil {
		return nil, err
	}
	return &Variant{
		Name:  id.String(),
		Value: v_,
		Type:  v,
	}, nil
}

func (v VariantType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[v.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", v)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (v VariantType) EncodeValue(value any) ([]byte, error) {
	fs, ok := value.(Variant)
	if !ok {
		return nil, NewEncodeValueError(v, varType)
	}
	for i, f := range v.Fields {
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

func (v VariantType) String() string {
	var s []string
	for _, f := range v.Fields {
		s = append(s, fmt.Sprintf("%s:%s", f.Name, f.Type.String()))
	}
	return fmt.Sprintf("variant {%s}", strings.Join(s, "; "))
}
