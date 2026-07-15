package idl

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/aviate-labs/agent-go/leb128"
)

// RecursiveType stands in for a composite type that (transitively) refers to
// itself. TypeOf installs one of these as a placeholder before expanding a
// struct's fields, so a self-reference resolves to the same node instead of
// recursing forever. Its String() is a stable name (not the expanded body),
// which both terminates RecordType.String() and lets the type-definition
// table dedupe the self-reference to a single index.
//
// All behaviour delegates to the resolved inner Type; RecursiveType is purely
// a naming/cycle-breaking indirection and adds nothing to the wire encoding
// beyond what its inner type produces.
type RecursiveType struct {
	name  string
	inner Type
	used  bool // set when a back-reference actually resolved to this placeholder
}

// NewRecursiveType creates an unresolved placeholder with the given name.
// Call setInner once the real type is built.
func NewRecursiveType(name string) *RecursiveType {
	return &RecursiveType{name: name}
}

func (r *RecursiveType) setInner(t Type) { r.inner = t }

func (r *RecursiveType) resolved() Type { return r.inner }

// Used reports whether this placeholder was referenced during type expansion,
// i.e. whether the type is genuinely recursive.
func (r *RecursiveType) Used() bool { return r.used }

func (r *RecursiveType) markUsed() { r.used = true }

func (r *RecursiveType) AddTypeDefinition(tdt *TypeDefinitionTable) error {
	if r.inner == nil {
		return nil
	}
	// Re-entry during the inner type's own definition: the placeholder is
	// already registered, so the self-reference is satisfied. Break the cycle.
	if _, ok := tdt.Indexes[r.String()]; ok {
		return nil
	}
	// Reserve a slot for this recursive type and register the placeholder name
	// against it BEFORE descending, so a field that refers back here resolves
	// to this index instead of recursing forever (candid forward reference).
	slot := len(tdt.Types)
	tdt.Types = append(tdt.Types, nil) // placeholder, filled in below
	tdt.Indexes[r.String()] = slot

	if err := r.inner.AddTypeDefinition(tdt); err != nil {
		return err
	}
	// The inner type appended its own definition (under its own String()).
	// Move those bytes into the reserved slot and point the inner's index at
	// it too, so both names resolve to one slot.
	innerIdx, ok := tdt.Indexes[r.inner.String()]
	if !ok {
		return fmt.Errorf("recursive inner type not defined: %s", r.inner)
	}
	tdt.Types[slot] = tdt.Types[innerIdx]
	tdt.Indexes[r.inner.String()] = slot
	return nil
}

func (r *RecursiveType) Decode(reader *bytes.Reader) (any, error) {
	return r.inner.Decode(reader)
}

func (r *RecursiveType) EncodeType(tdt *TypeDefinitionTable) ([]byte, error) {
	// Prefer the placeholder's reserved slot: during the inner record's own
	// definition a self-referencing field encodes before the record has
	// registered under its own String(), but the placeholder name is already
	// indexed.
	if idx, ok := tdt.Indexes[r.String()]; ok {
		return leb128.EncodeSigned(big.NewInt(int64(idx)))
	}
	return r.inner.EncodeType(tdt)
}

func (r *RecursiveType) EncodeValue(v any) ([]byte, error) {
	return r.inner.EncodeValue(v)
}

func (r *RecursiveType) UnmarshalGo(raw any, v any) error {
	return r.inner.UnmarshalGo(raw, v)
}

func (r *RecursiveType) Read(reader *bytes.Reader) ([]byte, error) {
	return r.inner.Read(reader)
}

func (r *RecursiveType) String() string {
	return r.name
}
