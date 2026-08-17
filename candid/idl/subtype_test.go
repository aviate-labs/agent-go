package idl_test

import (
	"testing"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
)

// type Opt = opt Opt has no finite value other than null, so promoting a bool
// into it never bottoms out.
func TestSubtypeUnboundedOpt(t *testing.T) {
	rec := idl.NewRecursiveType("Opt")
	rec.SetInner(idl.NewOptionalType(rec))
	if err := idl.Subtype(new(idl.BoolType), rec); err == nil {
		t.Fatal("expected bool <: Opt to fail")
	}
	if err := idl.Subtype(new(idl.NullType), rec); err != nil {
		t.Fatalf("null <: Opt: %v", err)
	}
}

// A decoded service can carry an unresolved signature; subtyping must report it
// rather than dereference nil.
func TestSubtypeServiceNilFunc(t *testing.T) {
	withFunc := &idl.ServiceType{Methods: []idl.Method{{Name: "m", Func: &idl.FunctionType{}}}}
	unresolved := &idl.ServiceType{Methods: []idl.Method{{Name: "m"}}}
	for _, tc := range []struct {
		name string
		a, b *idl.ServiceType
	}{
		{"nil on the left", unresolved, withFunc},
		{"nil on the right", withFunc, unresolved},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := idl.Subtype(tc.a, tc.b); err == nil {
				t.Fatal("expected an error for an unresolved signature")
			}
		})
	}
}

// A deep option chain is finite and must be accepted; only one with no end is
// rejected. A depth cap cannot tell them apart.
func TestSubtypeDeepOptChain(t *testing.T) {
	var deep idl.Type = new(idl.NatType)
	for range 70 {
		deep = idl.NewOptionalType(deep)
	}
	if err := idl.Subtype(new(idl.BoolType), deep); err != nil {
		t.Fatalf("70-deep opt chain rejected: %v", err)
	}

	rec := idl.NewRecursiveType("Opt")
	rec.SetInner(idl.NewOptionalType(rec))
	done := make(chan error, 1)
	go func() { done <- idl.Subtype(new(idl.BoolType), rec) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected `type Opt = opt Opt` to be rejected")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Subtype did not terminate")
	}
}
