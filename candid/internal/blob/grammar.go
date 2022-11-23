// Do not edit. This file is auto-generated.
// Grammar: CANDID-BLOB (v0.1.0) github.com/di-wu/candid-go/internal/blob

package blob

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
)

// Node Types
const (
	Unknown = iota

	// CANDID-BLOB (github.com/di-wu/candid-go/internal/blob)

	BlobT  // 001
	AlphaT // 002
	HexT   // 003
)

var NodeTypes = []string{
	"UNKNOWN",

	// CANDID-BLOB (github.com/di-wu/candid-go/internal/blob)

	"Blob",
	"Alpha",
	"Hex",
}

func Alpha(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        AlphaT,
			TypeStrings: NodeTypes,
			Value: op.MinOne(
				op.Or{
					parser.CheckRuneRange('A', 'Z'),
					parser.CheckRuneRange('a', 'z'),
				},
			),
		},
	)
}

func Blob(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BlobT,
			TypeStrings: NodeTypes,
			Value: op.MinOne(
				op.Or{
					Alpha,
					op.And{
						'\\',
						Hex,
					},
				},
			),
		},
	)
}

func Hex(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        HexT,
			TypeStrings: NodeTypes,
			Value: op.Repeat(2,
				op.Or{
					parser.CheckRuneRange('0', '9'),
					parser.CheckRuneRange('a', 'f'),
					parser.CheckRuneRange('A', 'F'),
				},
			),
		},
	)
}
