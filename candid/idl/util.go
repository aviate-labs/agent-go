package idl

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"

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

// DecodeLen reads a ULEB128 length and rejects values that cannot be honored by
// the remaining input, so a malformed header cannot make() a huge or negative
// allocation. Each element is at least one byte, so a length exceeding r.Len()
// is always invalid.
func DecodeLen(r *bytes.Reader) (int, error) {
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return 0, err
	}
	return checkLen(l, r)
}

const maxTypeValues = 1 << 16

// valueCount reports how many values one value of t expands into, saturating at
// maxTypeValues.
func valueCount(t Type, seen []Type) int {
	switch t := t.(type) {
	case *RecordType:
		if slices.Contains(seen, Type(t)) {
			return 1
		}
		seen = append(seen, t)
		n := 1
		for _, f := range t.Fields {
			if n += valueCount(f.Type, seen); n >= maxTypeValues {
				return maxTypeValues
			}
		}
		return n
	case *RecursiveType:
		if t.inner == nil || slices.Contains(seen, Type(t)) {
			return 1
		}
		return valueCount(t.inner, append(seen, t))
	case *VectorType:
		// One element: the count itself is bounded by the payload.
		return valueCount(t.Type, seen)
	case *OptionalType:
		return valueCount(t.Type, seen)
	case *FunctionType:
		// A func value is just a reference on the wire, but its signature is
		// still a path a cycle can run through, and String() walks it.
		if slices.Contains(seen, Type(t)) {
			return 1
		}
		seen = append(seen, t)
		n := 1
		for _, p := range slices.Concat(t.ArgumentParameters, t.ReturnParameters) {
			if p.Type == nil {
				continue
			}
			if n += valueCount(p.Type, seen); n >= maxTypeValues {
				return maxTypeValues
			}
		}
		return n
	case *ServiceType:
		if slices.Contains(seen, Type(t)) {
			return 1
		}
		seen = append(seen, t)
		n := 1
		for _, m := range t.Methods {
			if m.Func == nil {
				continue
			}
			if n += valueCount(m.Func, seen); n >= maxTypeValues {
				return maxTypeValues
			}
		}
		return n
	default:
		return 1
	}
}

// CheckSize rejects a type describing more values than the payload could ever
// justify: a record holding two copies of the one below it doubles per level,
// so twenty type-table entries can name a million values.
func CheckSize(t Type) error {
	if n := valueCount(t, nil); n >= maxTypeValues {
		return fmt.Errorf("type expands to at least %d values", n)
	}
	return nil
}

// decodeLenOf reads an element count. A zero-width element type carries no
// bytes, so the remaining input cannot bound the count and the zero-width
// ceiling is the only limit; anything else is still bounded by bytes remaining.
// The length is clamped before being returned as an int, so the caller's
// per-element charge cannot overflow.
func decodeLenOf(r *bytes.Reader, elem Type, zeroWidth bool, budget *Budget) (int, error) {
	if !zeroWidth {
		return DecodeLen(r)
	}
	l, err := leb128.DecodeUnsigned(r)
	if err != nil {
		return 0, err
	}
	if !l.IsInt64() || l.Int64() < 0 || l.Int64() > math.MaxInt32 {
		return 0, fmt.Errorf("invalid length %s", l)
	}
	if rem, bounded := budget.RemainingZeroWidth(); bounded && l.Int64() > int64(rem) {
		return 0, fmt.Errorf("length %s exceeds the zero-width limit", l)
	}
	return int(l.Int64()), nil
}

// isZeroWidth reports whether a value of t encodes to no bytes at all. Records
// may be cyclic, so track the ones already being considered.
func isZeroWidth(t Type) bool { return zeroWidth(t, nil) }

func zeroWidth(t Type, seen []Type) bool {
	switch t := t.(type) {
	case *NullType, *ReservedType:
		return true
	case *RecordType:
		if slices.Contains(seen, Type(t)) {
			return false
		}
		seen = append(seen, t)
		for _, f := range t.Fields {
			if !zeroWidth(f.Type, seen) {
				return false
			}
		}
		return true
	default:
		return false
	}
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

// cyclicStringer renders itself given the composites already being rendered
// further up the call stack. String() cannot pass that state through, so the
// composite types implement this alongside it and String() seeds the walk.
type cyclicStringer interface {
	stringSeen(seen []Type) string
	// elided renders this type where it encloses itself.
	elided() string
}

// typeString renders t, eliding a composite that encloses itself. Every
// composite is reached through a pointer, so a cycle is the same pointer twice
// on the path. Tracking the path rather than a depth keeps a deep-but-acyclic
// type intact, which matters because TypeDefinitionTable keys on String().
func typeString(t Type, seen []Type) string {
	c, ok := t.(cyclicStringer)
	if !ok {
		return t.String()
	}
	if slices.Contains(seen, t) {
		return c.elided()
	}
	return c.stringSeen(append(seen, t))
}
