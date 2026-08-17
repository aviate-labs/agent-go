package idl

import (
	"slices"
)

type TypeDefinitionTable struct {
	Types   [][]byte
	Indexes map[string]int

	// visiting is the AddTypeDefinition path currently being walked. A decoded
	// type can hold a raw structural cycle with no RecursiveType to break it,
	// and the walk would not otherwise terminate.
	visiting []Type
}

// enter marks t as being defined, reporting false if it already is. The returned
// function unmarks it.
func (tdt *TypeDefinitionTable) enter(t Type) (func(), bool) {
	if slices.Contains(tdt.visiting, t) {
		return nil, false
	}
	tdt.visiting = append(tdt.visiting, t)
	return func() { tdt.visiting = tdt.visiting[:len(tdt.visiting)-1] }, true
}

// NewTypeDefinitionTable returns a table ready to Add to: the zero value has a
// nil Indexes map, which Add writes to unconditionally.
func NewTypeDefinitionTable() *TypeDefinitionTable {
	return &TypeDefinitionTable{Indexes: make(map[string]int)}
}

func (tdt *TypeDefinitionTable) Add(t Type, bs []byte) {
	if i := slices.IndexFunc(tdt.Types, func(typ []byte) bool {
		return slices.Equal(typ, bs)
	}); i != -1 {
		tdt.Indexes[t.String()] = i
		return
	}

	i := len(tdt.Types)
	tdt.Indexes[t.String()] = i
	tdt.Types = append(tdt.Types, bs)
}
