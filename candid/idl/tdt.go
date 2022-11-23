package idl

type TypeDefinitionTable struct {
	Types   [][]byte
	Indexes map[string]int
}

func (tdt *TypeDefinitionTable) Add(t Type, bs []byte) {
	i := len(tdt.Types)
	tdt.Indexes[t.String()] = i
	tdt.Types = append(tdt.Types, bs)
}
