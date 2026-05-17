package leb128

import (
	"bytes"
	"fmt"
	"math/big"
)

// DecodeUnsigned converts the byte slice back to an unsigned integer.
func DecodeUnsigned(r *bytes.Reader) (*big.Int, error) {
	var (
		weight = big.NewInt(1)
		value  = new(big.Int)
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		value = value.Add(
			value,
			new(big.Int).Mul(big.NewInt(int64(b&0x7F)), weight),
		)
		weight = weight.Mul(weight, x80)
		if b < 0x80 {
			break
		}
	}
	return value, nil
}

// LEB128 represents an unsigned number encoded using (unsigned) LEB128.
type LEB128 []byte

// EncodeUnsigned encodes an unsigned integer.
func EncodeUnsigned(n *big.Int) (LEB128, error) {
	v := new(big.Int).Set(n)
	if v.Sign() < 0 {
		return nil, fmt.Errorf("can not leb128 encode negative values")
	}
	var bs []byte
	for {
		i := new(big.Int).And(v, x7F)
		v = v.Div(v, x80)
		if v.Cmp(x00) == 0 {
			b := i.Bytes()
			if len(b) == 0 {
				return []byte{0}, nil
			}
			return append(bs, b...), nil
		} else {
			b := new(big.Int).Or(i, x80)
			bs = append(bs, b.Bytes()...)
		}
	}
}
