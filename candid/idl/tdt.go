package idl

import (
	"slices"
)

type TypeDefinitionTable struct {
	Types   [][]byte
	Indexes map[string]int
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
