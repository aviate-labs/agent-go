package idl

import (
	"bytes"
	"fmt"
	"math/big"
)

type OpCode int64

var (
	NullOpCode      OpCode = -1  // 0x7f
	BoolOpCode      OpCode = -2  // 0x7e
	NatOpCode       OpCode = -3  // 0x7d
	IntOpCode       OpCode = -4  // 0x7c
	NatXOpCode      OpCode = -5  // 0x7b-0x78
	IntXOpCode      OpCode = -9  // 0x77-0x73
	FloatXOpCode    OpCode = -13 // 0x72
	TextOpCode      OpCode = -15 // 0x71
	ReservedOpCode  OpCode = -16 // 0x70
	EmptyOpCode     OpCode = -17 // 0x6f
	OptOpCode       OpCode = -18 // 0x6e
	VecOpCode       OpCode = -19 // 0x6d
	RecOpCode       OpCode = -20 // 0x6c
	VarOpCode       OpCode = -21 // 0x6b
	FuncOpCode      OpCode = -22 // 0x6a
	ServiceOpCode   OpCode = -23 // 0x69
	PrincipalOpCode OpCode = -24 // 0x68
)

func (o OpCode) BigInt() *big.Int {
	return big.NewInt(int64(o))
}

func (o OpCode) GetType(tds []Type) (Type, error) {
	if o >= 0 {
		if int(o) >= len(tds) || tds[o] == nil {
			return nil, fmt.Errorf("type index out of range: %d", o)
		}
		return tds[o], nil
	}

	switch o {
	case NullOpCode:
		return new(NullType), nil
	case BoolOpCode:
		return new(BoolType), nil
	case NatOpCode:
		return new(NatType), nil
	case IntOpCode:
		return new(IntType), nil
	case NatXOpCode:
		return Nat8Type(), nil
	case NatXOpCode - 1:
		return Nat16Type(), nil
	case NatXOpCode - 2:
		return Nat32Type(), nil
	case NatXOpCode - 3:
		return Nat64Type(), nil
	case IntXOpCode:
		return Int8Type(), nil
	case IntXOpCode - 1:
		return Int16Type(), nil
	case IntXOpCode - 2:
		return Int32Type(), nil
	case IntXOpCode - 3:
		return Int64Type(), nil
	case FloatXOpCode:
		return Float32Type(), nil
	case FloatXOpCode - 1:
		return Float64Type(), nil
	case TextOpCode:
		return new(TextType), nil
	case ReservedOpCode:
		return new(ReservedType), nil
	case EmptyOpCode:
		return new(EmptyType), nil
	case PrincipalOpCode:
		return new(PrincipalType), nil
	default:
		if o < -24 {
			return nil, &FormatError{
				Description: "type: out of range",
			}
		}
		return nil, &FormatError{
			Description: "type: not primitive",
		}
	}
}

func (o OpCode) String() string {
	switch o {
	case NullOpCode:
		return "null"
	case BoolOpCode:
		return "bool"
	case NatOpCode:
		return "nat"
	case IntOpCode:
		return "int"
	case NatXOpCode:
		return "nat8"
	case NatXOpCode - 1:
		return "nat16"
	case NatXOpCode - 2:
		return "nat32"
	case NatXOpCode - 3:
		return "nat64"
	case IntXOpCode:
		return "int8"
	case IntXOpCode - 1:
		return "int16"
	case IntXOpCode - 2:
		return "int32"
	case IntXOpCode - 3:
		return "int64"
	case FloatXOpCode:
		return "float32"
	case FloatXOpCode - 1:
		return "float64"
	case TextOpCode:
		return "text"
	case ReservedOpCode:
		return "reserved"
	case EmptyOpCode:
		return "empty"
	case OptOpCode:
		return "opt"
	case VecOpCode:
		return "vec"
	case RecOpCode:
		return "rec"
	case VecOpCode:
		return "var"
	case FuncOpCode:
		return "func"
	case ServiceOpCode:
		return "service"
	case PrincipalOpCode:
		return "principal"
	default:
		return "unknown"
	}
}

type PrimType interface {
	prim()
}

type Type interface {
	// AddTypeDefinition adds itself to the definition table if it is not a primitive type.
	AddTypeDefinition(*TypeDefinitionTable) error

	// Decode decodes the value from the reader.
	Decode(*bytes.Reader) (any, error)

	// EncodeType encodes the type.
	EncodeType(*TypeDefinitionTable) ([]byte, error)

	// EncodeValue encodes the value.
	EncodeValue(v any) ([]byte, error)

	// UnmarshalGo unmarshals the value from the go value.
	UnmarshalGo(raw any, v any) error

	fmt.Stringer
}

type primType struct{}

func (primType) AddTypeDefinition(_ *TypeDefinitionTable) error {
	return nil // No need to add primitive types to the type definition table.
}

func (primType) prim() {}
