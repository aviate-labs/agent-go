package ctest

import (
	"maps"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid/internal/candid"
)

// NewTestParser parses a `*.test.did` conformance file.
//
// The generated NewParser binds DataType to this package's placeholder, which
// only refers to itself. The type language in these files is the candid one, so
// bind the real rules here instead; the abnf generator cannot express a
// reference into another grammar.
func NewTestParser(input []rune) (*parser.Parser, error) {
	p, err := parser.New(input)
	if err != nil {
		return nil, err
	}
	maps.Copy(p.Rules, map[string]parser.Operator{
		"ActorType": candid.ActorType,
		"DataType":  candid.DataType,
		"FieldType": candid.FieldType,
		"Func":      candid.Func,
		"FuncType":  candid.FuncType,
		"Opt":       candid.Opt,
		"Service":   candid.Service,
		"Vec":       candid.Vec,
	})
	return p, nil
}
