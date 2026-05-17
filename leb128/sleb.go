package leb128

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
)

// DecodeSigned converts the byte slice back to a signed integer.
func DecodeSigned(r *bytes.Reader) (*big.Int, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	l := 0
	for _, b := range bs {
		if b < 0x80 {
			if (b & 0x40) == 0 {
				*r = *bytes.NewReader(bs)
				return DecodeUnsigned(r)
			}
			break
		}
		l++
	}
	if l >= len(bs) {
		return nil, fmt.Errorf("too short")
	}
	*r = *bytes.NewReader(bs[l+1:])

	v := new(big.Int)
	for i := l; i >= 0; i-- {
		v = v.Mul(v, x80)
		v = v.Add(v, big.NewInt(int64(0x80-(bs[i]&0x7F)-1)))
	}
	v = v.Mul(v, big.NewInt(-1))
	v = v.Add(v, big.NewInt(-1))
	return v, nil
}

// EncodeSigned encodes a signed integer.
func EncodeSigned(n *big.Int) (LEB128, error) {
	v := new(big.Int).Set(n)
	neg := v.Sign() < 0
	if neg {
		v = v.Mul(v, big.NewInt(-1))
		v = v.Add(v, big.NewInt(-1))
	}
	var bs []byte
	for {
		b := byte(v.Int64() % 0x80)
		if neg {
			b = 0x80 - b - 1
		}
		v = v.Div(v, x80)
		if (neg && v.Sign() == 0 && b&0x40 != 0) ||
			(!neg && v.Sign() == 0 && b&0x40 == 0) {
			return append(bs, b), nil
		} else {
			bs = append(bs, b|0x80)
		}
	}
}

// SLEB128 represents a signed number encoded using signed LEB128.
type SLEB128 []byte
