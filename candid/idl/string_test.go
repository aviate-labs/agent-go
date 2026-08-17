package idl_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func deepRecord(leaf idl.Type, depth int) idl.Type {
	t := leaf
	for range depth {
		t = idl.NewRecordType(map[string]idl.Type{"f": t})
	}
	return t
}

// TypeDefinitionTable keys on String(), so a deep but acyclic type must render
// in full: two types that stringify the same would share a type-table slot.
func TestStringDeepAcyclic(t *testing.T) {
	const depth = 64
	s := deepRecord(new(idl.NatType), depth).String()
	if strings.Contains(s, "...") {
		t.Fatalf("deep acyclic record truncated: %s", s)
	}
	if got := strings.Count(s, "record {"); got != depth {
		t.Fatalf("got %d levels, want %d", got, depth)
	}
	if a, b := deepRecord(new(idl.NatType), depth), deepRecord(new(idl.TextType), depth); a.String() == b.String() {
		t.Fatal("distinct deep types collide")
	}
}

// A cycle with no RecursiveType to break it still has to terminate.
func TestStringCyclic(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func() idl.Type
	}{
		{"record", func() idl.Type {
			r := &idl.RecordType{}
			r.Fields = []idl.FieldType{{Name: "a", Type: r}}
			return r
		}},
		{"variant", func() idl.Type {
			v := &idl.VariantType{}
			v.Fields = []idl.FieldType{{Name: "a", Type: v}}
			return v
		}},
		// The cycle runs through a vector, which must carry the path along.
		{"record via vec", func() idl.Type {
			r := &idl.RecordType{}
			r.Fields = []idl.FieldType{{Name: "a", Type: idl.NewVectorType(r)}}
			return r
		}},
		{"record via opt", func() idl.Type {
			r := &idl.RecordType{}
			r.Fields = []idl.FieldType{{Name: "a", Type: idl.NewOptionalType(r)}}
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			done := make(chan string, 1)
			go func() { done <- tc.make().String() }()
			if s := <-done; !strings.Contains(s, "...") {
				t.Fatalf("cycle not elided: %s", s)
			}
		})
	}
}

// Sibling references are not a cycle: the same type twice must render twice.
func TestStringSharedNotCycle(t *testing.T) {
	shared := idl.NewRecordType(map[string]idl.Type{"x": new(idl.NatType)})
	s := idl.NewRecordType(map[string]idl.Type{"a": shared, "b": shared}).String()
	if strings.Contains(s, "...") {
		t.Fatalf("shared non-cyclic type elided: %s", s)
	}
}

// String() is called from EncodeType via the type table, so it has to be safe
// to render the same type from several goroutines.
func TestStringConcurrent(t *testing.T) {
	r := deepRecord(new(idl.NatType), 16)
	want := r.String()
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if got := r.String(); got != want {
				t.Errorf("concurrent String() mismatch")
			}
		})
	}
	wg.Wait()
}

// Two funcs sharing one parameter slice are distinct, not a cycle: eliding one
// would collapse them into a single TypeDefinitionTable slot.
func TestStringSharedParamsNotCycle(t *testing.T) {
	shared := []idl.FunctionParameter{{Type: new(idl.NatType)}}
	inner := &idl.FunctionType{
		ArgumentParameters: shared,
		ReturnParameters:   []idl.FunctionParameter{{Type: new(idl.TextType)}},
	}
	outer := &idl.FunctionType{
		ArgumentParameters: []idl.FunctionParameter{{Type: inner}},
		ReturnParameters:   shared,
	}
	if s := outer.String(); strings.Contains(s, "func") {
		t.Fatalf("distinct funcs elided as a cycle: %s", s)
	}
}

// A func with no results is still identifiable by its arguments, so a cycle
// through one terminates.
func TestStringCyclicFuncNoResults(t *testing.T) {
	f := &idl.FunctionType{}
	f.ArgumentParameters = []idl.FunctionParameter{{Type: f}}
	done := make(chan string, 1)
	go func() { done <- f.String() }()
	select {
	case s := <-done:
		if !strings.Contains(s, "func") {
			t.Fatalf("cycle not elided: %s", s)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("String() did not terminate")
	}
}

// A decoded type can hold a raw structural cycle with no RecursiveType to break
// it. AddTypeDefinition walks the same graph as String() and must terminate too.
func TestAddTypeDefinitionCyclic(t *testing.T) {
	r := &idl.RecordType{}
	r.Fields = []idl.FieldType{{Name: "x", Type: idl.NewVectorType(r)}}
	done := make(chan error, 1)
	go func() { done <- r.AddTypeDefinition(idl.NewTypeDefinitionTable()) }()
	select {
	case <-done: // an error is fine; not terminating is not
	case <-time.After(5 * time.Second):
		t.Fatal("AddTypeDefinition did not terminate")
	}
}
