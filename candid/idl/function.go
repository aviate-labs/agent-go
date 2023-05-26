package idl

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
)

func encodeTypes(ts []Type, tdt *TypeDefinitionTable) ([]byte, error) {
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(ts))))
	if err != nil {
		return nil, err
	}
	var vs []byte
	for _, t := range ts {
		v, err := t.EncodeType(tdt)
		if err != nil {
			return nil, err
		}
		vs = append(vs, v...)
	}
	return concat(l, vs), nil
}

type FunctionType struct {
	ArgTypes    []Type
	RetTypes    []Type
	Annotations []string
}

func NewFunctionType(argumentTypes []Type, returnTypes []Type, annotations []string) *FunctionType {
	return &FunctionType{
		ArgTypes:    argumentTypes,
		RetTypes:    returnTypes,
		Annotations: annotations,
	}
}

func (f FunctionType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	for _, t := range f.ArgTypes {
		if err := t.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}
	for _, t := range f.RetTypes {
		if err := t.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}

	id, err := leb128.EncodeSigned(big.NewInt(funcType))
	if err != nil {
		return err
	}
	vsa, err := encodeTypes(f.ArgTypes, tdt)
	if err != nil {
		return err
	}
	vsr, err := encodeTypes(f.RetTypes, tdt)
	if err != nil {
		return err
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(f.Annotations))))
	if err != nil {
		return err
	}
	var vs []byte
	for _, t := range f.Annotations {
		switch t {
		case "query":
			vs = []byte{0x01}
		case "oneway":
			vs = []byte{0x02}
		default:
			return fmt.Errorf("invalid function annotation: %s", t)
		}
	}

	tdt.Add(f, concat(id, vsa, vsr, l, vs))
	return nil
}

func (f FunctionType) Decode(r *bytes.Reader) (any, error) {
	{
		bs := make([]byte, 2)
		n, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		if n != 2 || bs[0] != 0x01 || bs[1] != 0x01 {
			return nil, fmt.Errorf("invalid func reference: %d", bs)
		}
	}
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	pid := make([]byte, l.Int64())
	{
		n, err := r.Read(pid)
		if err != nil {
			return nil, err
		}
		if n != int(l.Int64()) {
			return nil, fmt.Errorf("invalid principal id: %d", pid)
		}
	}
	ml, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return nil, err
	}
	m := make([]byte, ml.Int64())
	{
		n, err := r.Read(pid)
		if err != nil {
			return nil, err
		}
		if n != int(l.Int64()) {
			return nil, fmt.Errorf("invalid method: %d", pid)
		}
	}
	return &PrincipalMethod{
		Principal: principal.Principal{Raw: pid},
		Method:    string(m),
	}, nil
}

func (f FunctionType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[f.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", f)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (f FunctionType) EncodeValue(v any) ([]byte, error) {
	pm, ok := v.(PrincipalMethod)
	if !ok {
		return nil, NewEncodeValueError(v, funcType)
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(pm.Principal.Raw))))
	if err != nil {
		return nil, err
	}
	lm, err := leb128.EncodeUnsigned(big.NewInt(int64(len(pm.Method))))
	if err != nil {
		return nil, err
	}
	return concat([]byte{0x01, 0x01}, l, pm.Principal.Raw, lm, []byte(pm.Method)), nil
}

func (f FunctionType) String() string {
	var args []string
	for _, t := range f.ArgTypes {
		args = append(args, t.String())
	}
	var rets []string
	for _, t := range f.RetTypes {
		rets = append(rets, t.String())
	}
	var ann string
	if len(f.Annotations) != 0 {
		ann = fmt.Sprintf(" %s", strings.Join(f.Annotations, " "))
	}
	return fmt.Sprintf("(%s) -> (%s)%s", strings.Join(args, ", "), strings.Join(rets, ", "), ann)
}

func (FunctionType) UnmarshalGo(raw any, _v any) error {
	return NewUnmarshalGoError(raw, _v)
}

type PrincipalMethod struct {
	Principal principal.Principal
	Method    string
}
