package certification

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"

	"github.com/aviate-labs/leb128"
	"github.com/fxamacker/cbor/v2"
)

// HashAny computes the hash of any value.
func HashAny(v any) ([32]byte, error) {
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
		bi := big.NewInt(int64(v))
		e, err := leb128.EncodeUnsigned(bi)
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(e), nil
	case uint64:
		bi := big.NewInt(int64(v))
		e, err := leb128.EncodeUnsigned(bi)
		if err != nil {
			return [32]byte{}, err
		}
		return sha256.Sum256(e), nil
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
	var hashes [][]byte
	for _, kv := range m {
		if kv.Value == nil {
			continue
		}

		keyHash := sha256.Sum256([]byte(kv.Key))
		valueHash, err := HashAny(kv.Value)
		if err != nil {
			return [32]byte{}, err
		}
		hashes = append(hashes, append(keyHash[:], valueHash[:]...))
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) == -1
	})
	return sha256.Sum256(bytes.Join(hashes, nil)), nil
}

type KeyValuePair struct {
	Key   string
	Value any
}
