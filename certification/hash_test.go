package certification

import (
	"bytes"
	"encoding/hex"
	"testing"
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
