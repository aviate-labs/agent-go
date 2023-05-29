package idl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/aviate-labs/leb128"
	"math/big"
)

func anyToUint16(v any) (uint16, bool) {
	switch v := v.(type) {
	case uint16:
		return v, true
	case uint8:
		return uint16(v), true
	default:
		return 0, false
	}
}

func anyToUint32(v any) (uint32, bool) {
	switch v := v.(type) {
	case uint32:
		return v, true
	case uint16:
		return uint32(v), true
	case uint8:
		return uint32(v), true
	default:
		return 0, false
	}
}

func anyToUint64(v any) (uint64, bool) {
	switch v := v.(type) {
	case uint64:
		return v, true
	case uint32:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	default:
		return 0, false
	}
}

// encodeNat16 convert the given value to an uint16.
// Accepts: `uint8`, `uint16`.
func encodeNat16(v any) (uint16, error) {
	if v, ok := v.(uint16); ok {
		return v, nil
	}
	v_, err := encodeNat8(v)
	return uint16(v_), err
}

// encodeNat32 convert the given value to an uint32.
// Accepts: `uint8`, `uint16`, `uint32`.
func encodeNat32(v any) (uint32, error) {
	if v, ok := v.(uint32); ok {
		return v, nil
	}
	v_, err := encodeNat16(v)
	return uint32(v_), err
}

// encodeNat64 convert the given value to an uint64.
// Accepts: `uint8`, `uint16`, `uint32`, `uint64`.
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

// encodeNat8 convert the given value to an uint8.
// Accepts: `uint8`.
func encodeNat8(v any) (uint8, error) {
	if v, ok := v.(uint8); ok {
		return v, nil
	}
	return 0, fmt.Errorf("invalid value: %v", v)
}

// Nat represents an unbounded natural number.
type Nat struct {
	n *big.Int
}

// NewBigNat creates a new Nat from a big.Int.
func NewBigNat(bi *big.Int) Nat {
	return Nat{bi}
}

// NewNat creates a new Nat from any unsigned integer.
func NewNat[number Natural](n number) Nat {
	return Nat{new(big.Int).SetUint64(uint64(n))}
}

// NewNatFromString creates a new Nat from a string.
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

func anyToNat(v any) (Nat, bool) {
	switch v := v.(type) {
	case Nat:
		return v, true
	case uint:
		return NewNat(v), true
	case uint64:
		return NewNat(v), true
	case uint32:
		return NewNat(v), true
	case uint16:
		return NewNat(v), true
	case uint8:
		return NewNat(v), true
	default:
		return Nat{}, false
	}
}

// BigInt returns the underlying big.Int.
func (n Nat) BigInt() *big.Int {
	return n.n
}

// String returns the string representation of the Nat.
func (n Nat) String() string {
	return n.n.String()
}

// NatType is either a type of nat8, nat16, nat32, nat64, or nat.
type NatType struct {
	size uint8
	primType
}

// Nat16Type returns a type of nat16.
func Nat16Type() *NatType {
	return &NatType{
		size: 2,
	}
}

// Nat32Type returns a type of nat32.
func Nat32Type() *NatType {
	return &NatType{
		size: 4,
	}
}

// Nat64Type returns a type of nat64.
func Nat64Type() *NatType {
	return &NatType{
		size: 8,
	}
}

// Nat8Type returns a type of nat8.
func Nat8Type() *NatType {
	return &NatType{
		size: 1,
	}
}

// Base returns the base type of the NatType.
func (n NatType) Base() uint {
	return uint(n.size)
}

// Decode decodes an unsigned integer from the given reader.
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

// EncodeType returns the leb128 encoding of the NatType.
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

// EncodeValue encodes an nat value.
// Accepts: `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `Nat`.
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
		return nil, NewEncodeValueError(v, natType)
	}
}

// String returns the string representation of the NatType.
func (n NatType) String() string {
	if n.size == 0 {
		return "nat"
	}
	return fmt.Sprintf("nat%d", n.size*8)
}

func (n NatType) UnmarshalGo(raw any, _v any) error {
	switch n.size {
	case 0:
		n, ok := anyToNat(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		v, ok := _v.(*Nat)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		*v = n
		return nil
	case 8:
		u64, ok := anyToUint64(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *uint64:
			*v = u64
			return nil
		case *Nat:
			*v = NewNat(u64)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 4:
		u32, ok := anyToUint32(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *uint32:
			*v = u32
			return nil
		case *Nat:
			*v = NewNat(u32)
			return nil
		case *uint64:
			*v = uint64(u32)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 2:
		u16, ok := anyToUint16(raw)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *uint16:
			*v = u16
			return nil
		case *Nat:
			*v = NewNat(u16)
			return nil
		case *uint64:
			*v = uint64(u16)
			return nil
		case *uint32:
			*v = uint32(u16)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	case 1:
		u8, ok := raw.(uint8)
		if !ok {
			return NewUnmarshalGoError(raw, _v)
		}
		switch v := _v.(type) {
		case *uint8:
			*v = u8
			return nil
		case *uint64:
			*v = uint64(u8)
			return nil
		case *uint32:
			*v = uint32(u8)
			return nil
		case *uint16:
			*v = uint16(u8)
			return nil
		case *Nat:
			*v = NewNat(u8)
			return nil
		default:
			return NewUnmarshalGoError(raw, _v)
		}
	default:
		return NewUnmarshalGoError(raw, _v)
	}
}

// Natural contains all unsigned integer types.
type Natural interface {
	uint | uint64 | uint32 | uint16 | uint8
}
