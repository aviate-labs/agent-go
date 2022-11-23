package idl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/aviate-labs/leb128"
)

func encodeNat16(v any) (uint16, error) {
	if v, ok := v.(uint16); ok {
		return v, nil
	}
	v_, err := encodeNat8(v)
	return uint16(v_), err
}

func encodeNat32(v any) (uint32, error) {
	if v, ok := v.(uint32); ok {
		return v, nil
	}
	v_, err := encodeNat16(v)
	return uint32(v_), err
}

func encodeNat64(v any) (uint64, error) {
	if v, ok := v.(uint); ok {
		return uint64(v), nil
	}
	if v, ok := v.(uint64); ok {
		return v, nil
	}
	v_, err := encodeNat16(v)
	return uint64(v_), err
}

func encodeNat8(v any) (uint8, error) {
	if v, ok := v.(uint8); ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid value: %v", v)
}

type Nat struct {
	n *big.Int
}

func NewBigNat(bi *big.Int) Nat {
	return Nat{bi}
}

func NewNat[number Natural](n number) Nat {
	return Nat{new(big.Int).SetUint64(uint64(n))}
}

func NewNatFromString(n string) Nat {
	bi, ok := new(big.Int).SetString(n, 10)
	if !ok {
		panic("number: invalid string: " + n)
	}
	if bi.Sign() < 0 {
		panic("number: negative nat")
	}
	return Nat{bi}
}

func (n Nat) BigInt() *big.Int {
	return n.n
}

func (n Nat) String() string {
	return n.n.String()
}

type NatType struct {
	size uint8
	primType
}

func Nat16Type() *NatType {
	return &NatType{
		size: 2,
	}
}

func Nat32Type() *NatType {
	return &NatType{
		size: 4,
	}
}

func Nat64Type() *NatType {
	return &NatType{
		size: 8,
	}
}

func Nat8Type() *NatType {
	return &NatType{
		size: 1,
	}
}

func (n NatType) Base() uint {
	return uint(n.size)
}

func (n NatType) Decode(r *bytes.Reader) (any, error) {
	switch n.size {
	case 0:
		bi, err := leb128.DecodeUnsigned(r)
		if err != nil {
			return nil, err
		}
		return NewBigNat(bi), nil
	case 8:
		v := make([]byte, 8)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 8 {
			return nil, fmt.Errorf("nat64: too short")
		}
		return binary.LittleEndian.Uint64(v), nil
	case 4:
		v := make([]byte, 4)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 4 {
			return nil, fmt.Errorf("nat32: too short")
		}
		return binary.LittleEndian.Uint32(v), nil
	case 2:
		v := make([]byte, 2)
		n, err := r.Read(v)
		if err != nil {
			return nil, err
		}
		if n != 2 {
			return nil, fmt.Errorf("nat16: too short")
		}
		return binary.LittleEndian.Uint16(v), nil
	case 1:
		return r.ReadByte()
	default:
		return nil, fmt.Errorf("invalid int type with size %d", n.size)
	}
}

func (n NatType) EncodeType(_ *TypeDefinitionTable) ([]byte, error) {
	if n.size == 0 {
		return leb128.EncodeSigned(big.NewInt(natType))
	}
	natXType := new(big.Int).Set(big.NewInt(natXType))
	natXType = natXType.Add(
		natXType,
		big.NewInt(3-int64(log2(n.size*8))),
	)
	return leb128.EncodeSigned(natXType)
}

func (n NatType) EncodeValue(v any) ([]byte, error) {
	switch n.size {
	case 0:
		v, ok := v.(Nat)
		if !ok {
			return nil, fmt.Errorf("invalid value: %v", v)
		}
		return leb128.EncodeUnsigned(v.BigInt())
	case 8:
		v, err := encodeNat64(v)
		if err != nil {
			return nil, err
		}
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, v)
		return bs, nil
	case 4:
		v, err := encodeNat32(v)
		if err != nil {
			return nil, err
		}
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, v)
		return bs, nil
	case 2:
		v, err := encodeNat16(v)
		if err != nil {
			return nil, err
		}
		bs := make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, v)
		return bs, nil
	case 1:
		v, err := encodeNat8(v)
		if err != nil {
			return nil, err
		}
		return []byte{v}, nil
	default:
		return nil, fmt.Errorf("invalid argument: %v", v)
	}
}

func (n NatType) String() string {
	if n.size == 0 {
		return "nat"
	}
	return fmt.Sprintf("nat%d", n.size*8)
}

type Natural interface {
	uint | uint64 | uint32 | uint16 | uint8
}
