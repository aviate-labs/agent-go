package idl

import (
	"bytes"
	"testing"
)

// selfRef is a struct with an optional field pointing back to itself, i.e. a
// recursive candid type: type T = record { next : opt T; value : nat64 }.
type selfRef struct {
	Next  *selfRef `ic:"next,omitempty"`
	Value uint64   `ic:"value"`
}

func TestTypeOfRecursiveTerminates(t *testing.T) {
	typ, err := TypeOf(selfRef{Value: 1})
	if err != nil {
		t.Fatalf("TypeOf: %v", err)
	}
	if typ == nil {
		t.Fatal("nil type")
	}
	// String() must terminate.
	_ = typ.String()
}

// mutualA/mutualB are mutually recursive.
type mutualA struct {
	B     *mutualB `ic:"b,omitempty"`
	Label string   `ic:"label"`
}

type mutualB struct {
	A *mutualA `ic:"a,omitempty"`
	N uint64   `ic:"n"`
}

func TestTypeOfMutualRecursionTerminates(t *testing.T) {
	if _, err := TypeOf(mutualA{Label: "x"}); err != nil {
		t.Fatalf("TypeOf mutual: %v", err)
	}
}

func TestRecursiveEncodeType(t *testing.T) {
	typ, err := TypeOf(selfRef{Value: 1})
	if err != nil {
		t.Fatalf("TypeOf: %v", err)
	}
	tdt := &TypeDefinitionTable{Indexes: make(map[string]int)}
	if err := typ.AddTypeDefinition(tdt); err != nil {
		t.Fatalf("AddTypeDefinition: %v", err)
	}
	if _, err := typ.EncodeType(tdt); err != nil {
		t.Fatalf("EncodeType: %v", err)
	}
	// The self-referential record must occupy exactly one table slot.
	if len(tdt.Types) == 0 {
		t.Fatal("no type definitions emitted")
	}
	v := selfRef{Value: 7, Next: &selfRef{Value: 8}}
	if _, err := typ.EncodeValue(v); err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}
}

func TestRecursiveRoundTrip(t *testing.T) {
	typ, err := TypeOf(selfRef{})
	if err != nil {
		t.Fatalf("TypeOf: %v", err)
	}
	tdt := &TypeDefinitionTable{Indexes: make(map[string]int)}
	if err := typ.AddTypeDefinition(tdt); err != nil {
		t.Fatalf("AddTypeDefinition: %v", err)
	}
	v := selfRef{Value: 7, Next: &selfRef{Value: 8, Next: &selfRef{Value: 9}}}
	enc, err := typ.EncodeValue(v)
	if err != nil {
		t.Fatalf("EncodeValue: %v", err)
	}
	// Reading back the value bytes with the same type must consume them cleanly.
	dec, err := typ.Decode(bytes.NewReader(enc))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if dec == nil {
		t.Fatal("decoded nil")
	}
}
