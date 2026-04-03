package certification

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/fxamacker/cbor/v2"
	"github.com/niccolofant/agent-go/candid/idl"
	"github.com/niccolofant/agent-go/leb128"
)

// HashAny computes the hash of any value.
func HashAny(v any) ([32]byte, error) {
	if hasher, ok := v.(Hasher); ok {
		return hasher.HashAny()
	}

	switch v := v.(type) {
	case cbor.RawMessage:
		var anyValue any
		if err := cbor.Unmarshal(v, &anyValue); err != nil {
			panic(err)
		}
		return HashAny(anyValue)
	case string:
		return sha256.Sum256([]byte(v)), nil
	case []byte:
		return sha256.Sum256(v), nil
	case int64:
		var buf [10]byte
		return sha256.Sum256(leb128.AppendSignedInt64(buf[:0], v)), nil
	case uint64:
		var buf [10]byte
		return sha256.Sum256(leb128.AppendUnsignedUint64(buf[:0], v)), nil
	case idl.Int:
		bi := v.BigInt()
		e, err := leb128.EncodeSigned(bi)
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(e), nil
	case idl.Nat:
		bi := v.BigInt()
		e, err := leb128.EncodeUnsigned(bi)
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(e), nil
	case leb128.SLEB128:
		return sha256.Sum256(v), nil
	case leb128.LEB128:
		return sha256.Sum256(v), nil
	case map[any]any: // cbor maps are not guaranteed to have string keys
		kv := make([]KeyValuePair, len(v))
		i := 0
		for k, v := range v {
			s, isString := k.(string)
			if !isString {
				return [32]byte{}, fmt.Errorf("unsupported type %T", k)
			}
			kv[i] = KeyValuePair{Key: s, Value: v}
			i++
		}
		return RepresentationIndependentHash(kv)
	case map[string]any:
		m := make([]KeyValuePair, len(v))
		i := 0
		for k, v := range v {
			m[i] = KeyValuePair{Key: k, Value: v}
			i++
		}
		return RepresentationIndependentHash(m)
	case []any:
		var hashes []byte
		for _, v := range v {
			valueHash, err := HashAny(v)
			if err != nil {
				return [32]byte{}, err
			}
			hashes = append(hashes, valueHash[:]...)
		}
		return sha256.Sum256(hashes), nil
	default:
		return [32]byte{}, fmt.Errorf("unsupported type %T", v)
	}
}

// RepresentationIndependentHash computes the hash of a map in a representation-independent way.
// https://internetcomputer.org/docs/current/references/ic-interface-spec/#hash-of-map
func RepresentationIndependentHash(m []KeyValuePair) ([32]byte, error) {
	// Each pair contributes one row of sha256(key) || sha256(value).
	type row [64]byte
	rows := make([]row, 0, len(m))
	for _, kv := range m {
		if kv.Value == nil {
			continue
		}

		keyHash := sha256.Sum256([]byte(kv.Key))
		valueHash, err := HashAny(kv.Value)
		if err != nil {
			return [32]byte{}, err
		}
		var r row
		copy(r[:32], keyHash[:])
		copy(r[32:], valueHash[:])
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		return bytes.Compare(rows[i][:], rows[j][:]) == -1
	})
	h := sha256.New()
	for i := range rows {
		h.Write(rows[i][:])
	}
	var out [32]byte
	h.Sum(out[:0])
	return out, nil
}

// Hasher is an interface for types that can hash any value.
type Hasher interface {
	HashAny() ([32]byte, error)
}

type KeyValuePair struct {
	Key   string
	Value any
}
