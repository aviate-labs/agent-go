package idl

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"

	"github.com/aviate-labs/agent-go/leb128"
	"github.com/aviate-labs/agent-go/principal"
)

type Method struct {
	Name string
	Func *FunctionType
}

type ServiceType struct {
	Methods []Method
}

func NewServiceType(methods map[string]*FunctionType) *ServiceType {
	var service ServiceType
	for k, v := range methods {
		service.Methods = append(service.Methods, Method{
			Name: k,
			Func: v,
		})
	}
	sort.Slice(service.Methods, func(i, j int) bool {
		return Hash(service.Methods[i].Name).Cmp(Hash(service.Methods[j].Name)) < 0
	})
	return &service
}

func (s ServiceType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	for _, f := range s.Methods {
		if err := f.Func.AddTypeDefinition(tdt); err != nil {
			return err
		}
	}

	id, err := leb128.EncodeSigned(ServiceOpCode.BigInt())
	if err != nil {
		return err
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(s.Methods))))
	if err != nil {
		return err
	}
	var vs []byte
	for _, f := range s.Methods {
		id := []byte(f.Name)
		l, err := leb128.EncodeUnsigned(big.NewInt(int64(len((id)))))
		if err != nil {
			return nil
		}
		t, err := f.Func.EncodeType(tdt)
		if err != nil {
			return nil
		}
		vs = concat(vs, l, id, t)
	}

	tdt.Add(s, concat(id, l, vs))
	return nil
}

func (s ServiceType) Decode(r *bytes.Reader, budget *Budget) (any, error) {
	if err := budget.Spend(costValue); err != nil {
		return nil, err
	}
	{
		bs := make([]byte, 1)
		n, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		if n != 1 || bs[0] != 0x01 {
			return nil, fmt.Errorf("invalid func reference: %d", bs)
		}
	}
	l, err := DecodeLen(r)
	if err != nil {
		return nil, err
	}
	pid := make([]byte, l)
	if _, err := io.ReadFull(r, pid); err != nil {
		return nil, err
	}
	return &principal.Principal{Raw: pid}, nil
}

func (s ServiceType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	idx, ok := tdt.Indexes[s.String()]
	if !ok {
		return nil, fmt.Errorf("missing type index for: %s", s)
	}
	return leb128.EncodeSigned(big.NewInt(int64(idx)))
}

func (s ServiceType) EncodeValue(v any) ([]byte, error) {
	p, ok := v.(principal.Principal)
	if !ok {
		return nil, NewEncodeValueError(v, ServiceOpCode)
	}
	l, err := leb128.EncodeUnsigned(big.NewInt(int64(len(p.Raw))))
	if err != nil {
		return nil, err
	}
	return concat([]byte{0x01}, l, []byte(p.Raw)), nil
}

func (s ServiceType) Read(r *bytes.Reader) ([]byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
		return nil, fmt.Errorf("invalid func reference: %d", b)
	}
	raw, err := readLEB128(r)
	if err != nil {
		return nil, err
	}
	lbi, err := leb128.DecodeUnsigned(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	l, err := checkLen(lbi, r)
	if err != nil {
		return nil, err
	}
	pid := make([]byte, l)
	{
		n, err := r.Read(pid)
		if err != nil {
			return nil, err
		}
		if n != l {
			return nil, fmt.Errorf("invalid principal id: %d", pid)
		}
	}
	return concat([]byte{b}, raw, pid), nil
}

func (s ServiceType) String() string {
	return typeString(&s, nil)
}

func (s *ServiceType) elided() string { return "service" }

// stringSeen keeps the enclosing composites in view: a cycle may run through a
// method signature, and %s on one would start a fresh walk that never ends.
func (s *ServiceType) stringSeen(seen []Type) string {
	var methods []string
	for _, m := range s.Methods {
		if m.Func == nil {
			methods = append(methods, fmt.Sprintf("%s:?", m.Name))
			continue
		}
		methods = append(methods, fmt.Sprintf("%s:%s", m.Name, typeString(m.Func, seen)))
	}
	return fmt.Sprintf("service {%s}", strings.Join(methods, "; "))
}

func (ServiceType) UnmarshalGo(raw any, _v any) error {
	return NewUnmarshalGoError(raw, _v)
}
