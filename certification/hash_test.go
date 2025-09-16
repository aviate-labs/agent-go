package certification

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func TestHashAny(t *testing.T) {
	for _, test := range []struct {
		name string
		v    any
		want string
	}{
		{
			name: "array",
			v:    []any{"a"},
			want: "bf5d3affb73efd2ec6c36ad3112dd933efed63c4e1cbffcfa88e2759c144f2d8",
		},
		{
			name: "array",
			v:    []any{"a", "b"},
			want: "e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
		},
		{
			name: "array",
			v:    []any{[]byte{97}, "b"},
			want: "e5a01fee14e0ed5c48714f22180f25ad8365b53f9779f79dc4a3d7e93963f94a",
		},
		{
			name: "array of arrays",
			v:    []any{[]any{"a"}},
			want: "eb48bdfa15fc43dbea3aabb1ee847b6e69232c0f0d9705935e50d60cce77877f",
		},
		{
			name: "array of arrays",
			v:    []any{[]any{"a", "b"}},
			want: "029fd80ca2dd66e7c527428fc148e812a9d99a5e41483f28892ef9013eee4a19",
		},
		{
			name: "array of arrays",
			v:    []any{[]any{"a", "b"}, []byte{97}},
			want: "aec3805593d9ec6df50da070597f73507050ce098b5518d0456876701ada7bb7",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := HashAny(test.v)
			if err != nil {
				t.Fatal(err)
			}
			want, err := hex.DecodeString(test.want)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got[:], want) {
				t.Fatalf("got %x, want %x", got, test.want)
			}
		})
	}
}

func TestHashAny_ICRC3(t *testing.T) {
	for _, test := range []struct {
		v        any
		expected string
	}{
		{
			v:        ICRC3Value{Nat: func() *idl.Nat { n := idl.NewNat(uint64(42)); return &n }()},
			expected: "684888c0ebb17f374298b65ee2807526c066094c701bcc7ebbe1c1095f494fc1",
		},
		{
			v:        ICRC3Value{Int: func() *idl.Int { i := idl.NewInt(int64(-42)); return &i }()},
			expected: "de5a6f78116eca62d7fc5ce159d23ae6b889b365a1739ad2cf36f925a140d0cc",
		},
		{
			v:        ICRC3Value{Text: func() *string { s := "Hello, World!"; return &s }()},
			expected: "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		},
		{
			v:        ICRC3Value{Blob: &[]byte{0x01, 0x02, 0x03, 0x04}},
			expected: "9f64a747e1b97f131fabb6b447296c9b6f0201e79fb3c5356e6c77e89b6a806a",
		},
		{
			v: ICRC3Value{Array: &[]ICRC3Value{
				{Nat: func() *idl.Nat { n := idl.NewNat(uint64(3)); return &n }()},
				{Text: func() *string { s := "foo"; return &s }()},
				{Blob: &[]byte{0x05, 0x06}},
			}},
			expected: "514a04011caa503990d446b7dec5d79e19c221ae607fb08b2848c67734d468d6",
		},
		{
			v: ICRC3Value{Map: &[]MapEntry{
				{
					Field0: "from",
					Field1: ICRC3Value{Blob: &[]byte{0x00, 0xab, 0xcd, 0xef, 0x00, 0x12, 0x34, 0x00, 0x56, 0x78, 0x9a, 0x00, 0xbc, 0xde, 0xf0, 0x00, 0x01, 0x23, 0x45, 0x67, 0x89, 0x00, 0xab, 0xcd, 0xef, 0x01}},
				},
				{
					Field0: "to",
					Field1: ICRC3Value{Blob: &[]byte{0x00, 0xab, 0x0d, 0xef, 0x00, 0x12, 0x34, 0x00, 0x56, 0x78, 0x9a, 0x00, 0xbc, 0xde, 0xf0, 0x00, 0x01, 0x23, 0x45, 0x67, 0x89, 0x00, 0xab, 0xcd, 0xef, 0x01}},
				},
				{
					Field0: "amount",
					Field1: ICRC3Value{Nat: func() *idl.Nat { n := idl.NewNat(uint64(42)); return &n }()},
				},
				{
					Field0: "created_at",
					Field1: ICRC3Value{Nat: func() *idl.Nat { n := idl.NewNat(uint64(1699218263)); return &n }()},
				},
				{
					Field0: "memo",
					Field1: ICRC3Value{Nat: func() *idl.Nat { n := idl.NewNat(uint64(0)); return &n }()},
				},
			}},
			expected: "c56ece650e1de4269c5bdeff7875949e3e2033f85b2d193c2ff4f7f78bdcfc75",
		},
	} {
		got, err := HashAny(test.v)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := hex.DecodeString(test.expected)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got[:], expected) {
			t.Fatalf("got %x, want %x", got, expected)
		}
	}
}

func TestRepresentationIndependentHash(t *testing.T) {
	for _, test := range []struct {
		name string
		kv   []KeyValuePair
		want string
	}{
		{
			name: "key-value map",
			kv: []KeyValuePair{
				{Key: "name", Value: "foo"},
				{Key: "message", Value: "Hello World!"},
				{Key: "answer", Value: uint64(42)},
			},
			want: "b0c6f9191e37dceafdfc47fbfc7e9cc95f21c7b985c2f7ba5855015c2a8f13ac",
		},
		{
			name: "duplicate keys",
			kv: []KeyValuePair{
				{Key: "name", Value: "foo"},
				{Key: "name", Value: "bar"},
				{Key: "message", Value: "Hello World!"},
			},
			want: "435f77c9bdeca5dba4a4b8a34e4f732b4311f1fc252ec6d4e8ee475234b170f9",
		},
		{
			name: "reordered keys",
			kv: []KeyValuePair{
				{Key: "name", Value: "bar"},
				{Key: "message", Value: "Hello World!"},
				{Key: "name", Value: "foo"},
			},
			want: "435f77c9bdeca5dba4a4b8a34e4f732b4311f1fc252ec6d4e8ee475234b170f9",
		},
		{
			name: "bytes",
			kv: []KeyValuePair{
				{Key: "bytes", Value: []byte{0x01, 0x02, 0x03, 0x04}},
			},
			want: "546729666d96a712bd94f902a0388e33f9a19a335c35bc3d95b0221a4a574455",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := RepresentationIndependentHash(test.kv)
			if err != nil {
				t.Fatal(err)
			}
			want, err := hex.DecodeString(test.want)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got[:], want) {
				t.Fatalf("got %x, want %x", got, test.want)
			}
		})
	}
}

type ICRC3Value struct {
	Blob  *[]byte       `ic:"Blob,variant" json:"Blob,omitempty"`
	Text  *string       `ic:"Text,variant" json:"Text,omitempty"`
	Nat   *idl.Nat      `ic:"Nat,variant" json:"Nat,omitempty"`
	Int   *idl.Int      `ic:"Int,variant" json:"Int,omitempty"`
	Array *[]ICRC3Value `ic:"Array,variant" json:"Array,omitempty"`
	Map   *[]MapEntry   `ic:"Map,variant" json:"Map,omitempty"`
}

func (v ICRC3Value) HashAny() ([32]byte, error) {
	switch {
	case v.Blob != nil:
		return HashAny(*v.Blob)
	case v.Text != nil:
		return HashAny(*v.Text)
	case v.Nat != nil:
		return HashAny(*v.Nat)
	case v.Int != nil:
		return HashAny(*v.Int)
	case v.Array != nil:
		arr := make([]any, len(*v.Array))
		for i, e := range *v.Array {
			arr[i] = e
		}
		return HashAny(arr)
	case v.Map != nil:
		kv := make([]KeyValuePair, len(*v.Map))
		for i, e := range *v.Map {
			kv[i] = KeyValuePair{Key: e.Field0, Value: e.Field1}
		}
		return RepresentationIndependentHash(kv)
	default:
		return [32]byte{}, fmt.Errorf("empty variant")
	}
}

type MapEntry struct {
	Field0 string     `ic:"0,tuple" json:"0"`
	Field1 ICRC3Value `ic:"1,tuple" json:"1"`
}
