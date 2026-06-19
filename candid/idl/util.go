package idl

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"

	"github.com/aviate-labs/agent-go/leb128"
)

func checkIsPtr(_v any) (reflect.Value, bool) {
	v := reflect.ValueOf(_v)
	if v.Kind() != reflect.Pointer {
		return v, false
	}
	v = v.Elem()
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, true
}

// checkLen validates an already-decoded length against the bytes remaining in
// r, for callers that read the length from a separate buffer.
func checkLen(l *big.Int, r *bytes.Reader) (int, error) {
	if !l.IsInt64() || l.Int64() < 0 || l.Int64() > int64(r.Len()) {
		return 0, fmt.Errorf("invalid length %s with %d bytes remaining", l, r.Len())
	}
	return int(l.Int64()), nil
}

func concat(bs ...[]byte) []byte {
	var l int
	for _, b := range bs {
		l += len(b)
	}
	tmp := make([]byte, l)
	var i int
	for _, b := range bs {
		i += copy(tmp[i:], b)
	}
	return tmp
}

// decodeLen reads a ULEB128 length and rejects values that cannot be honored by
// the remaining input, so a malformed header cannot panic make() with a
// negative or absurd length. Each element is at least one byte, so a length
// exceeding r.Len() is always invalid.
func decodeLen(r *bytes.Reader) (int, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return 0, err
	}
	return checkLen(l, r)
}

func log2(n uint8) uint8 {
	return uint8(math.Log2(float64(n)))
}

func pad0(n int, bs []byte) []byte {
	if len(bs) >= n {
		return bs
	}
	out := make([]byte, n)
	copy(out, bs)
	return out
}

func pad1(n int, bs []byte) []byte {
	if len(bs) >= n {
		return bs
	}
	out := make([]byte, n)
	copy(out, bs)
	for i := len(bs); i < n; i++ {
		out[i] = 0xff
	}
	return out
}

func readInt(bi *big.Int, n int) (*big.Int, error) {
	m := big.NewInt(2)
	m = m.Exp(m, big.NewInt(int64((n-1)*8+7)), nil)
	if bi.Cmp(m) >= 0 {
		v := new(big.Int).Set(m)
		v = v.Mul(v, big.NewInt(-2))
		bi = bi.Add(bi, v)
	}
	return bi, nil
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func twosCompl(bi *big.Int) *big.Int {
	inv := bi.Bytes()
	for i, b := range inv {
		inv[i] = ^b
	}
	bi.SetBytes(inv)
	return bi.Add(bi, big.NewInt(1))
}

func writeInt(bi *big.Int, n int) []byte {
	switch bi.Sign() {
	case 0:
		return zeros(n)
	case -1:
		bi := new(big.Int).Set(bi)
		return pad1(n, reverse(twosCompl(bi).Bytes()))
	default:
		return pad0(n, reverse(bi.Bytes()))
	}
}

func zeros(n int) []byte {
	return make([]byte, n)
}
