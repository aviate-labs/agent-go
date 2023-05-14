// Do not edit. This file is auto-generated.
// Grammar: CANDID (v0.1.0) github.com/di-wu/candid-go/internal/candid/candidvalue

package candidvalue

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
)

// Node Types
const (
	Unknown = iota

	// CANDID (github.com/di-wu/candid-go/internal/candid/candidvalue)

	ValuesT      // 001
	OptValueT    // 002
	NumT         // 003
	NumValueT    // 004
	NumTypeT     // 005
	BoolValueT   // 006
	BlobT        // 007
	NullT        // 008
	PrincipalT   // 009
	TextT        // 010
	TextValueT   // 011
	RecordT      // 012
	RecordFieldT // 013
	VariantT     // 014
	VecT         // 015
	IdT          // 016
)

// Token Definitions
const (
	// CANDID (github.com/di-wu/candid-go/internal/candid/candidvalue)

	ESC = 0x005C // \
)

var NodeTypes = []string{
	"UNKNOWN",

	// CANDID (github.com/di-wu/candid-go/internal/candid/candidvalue)

	"Values",
	"OptValue",
	"Num",
	"NumValue",
	"NumType",
	"BoolValue",
	"Blob",
	"Null",
	"Principal",
	"Text",
	"TextValue",
	"Record",
	"RecordField",
	"Variant",
	"Vec",
	"Id",
}

func Ascii(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		parser.CheckRuneRange(0x0020, 0x0021),
		parser.CheckRuneRange(0x0023, 0x005B),
		parser.CheckRuneRange(0x005D, 0x007E),
	})
}

func Blob(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BlobT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"blob",
				Spp,
				'"',
				op.MinZero(
					op.Repeat(2,
						Hex,
					),
				),
				'"',
			},
		},
	)
}

func Bool(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			BoolValue,
			op.Optional(
				op.And{
					Sp,
					':',
					Sp,
					"bool",
				},
			),
		},
	)
}

func BoolValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BoolValueT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				"true",
				"false",
			},
		},
	)
}

func Char(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Utf,
			op.And{
				ESC,
				op.Repeat(2,
					Hex,
				),
			},
			op.And{
				ESC,
				Escape,
			},
			op.And{
				"\\u{",
				HexNum,
				'}',
			},
		},
	)
}

func Digit(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(parser.CheckRuneRange('0', '9'))
}

func Escape(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		'n',
		'r',
		't',
		ESC,
		0x0022,
		0x0027,
	})
}

func Hex(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		Digit,
		parser.CheckRuneRange('A', 'F'),
		parser.CheckRuneRange('a', 'f'),
	})
}

func HexNum(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			Hex,
			op.MinZero(
				op.And{
					op.Optional(
						'_',
					),
					Hex,
				},
			),
		},
	)
}

func Id(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        IdT,
			TypeStrings: NodeTypes,
			Value: op.And{
				op.Or{
					Letter,
					'_',
				},
				op.MinZero(
					op.Or{
						Letter,
						Digit,
						'_',
					},
				),
			},
		},
	)
}

func Letter(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		parser.CheckRuneRange('A', 'Z'),
		parser.CheckRuneRange('a', 'z'),
	})
}

func Null(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NullT,
			TypeStrings: NodeTypes,
			Value:       "null",
		},
	)
}

func Num(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NumT,
			TypeStrings: NodeTypes,
			Value: op.And{
				NumValue,
				op.Optional(
					op.And{
						Sp,
						':',
						Sp,
						NumType,
					},
				),
			},
		},
	)
}

func NumType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NumTypeT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				"nat8",
				"nat16",
				"nat32",
				"nat64",
				"nat",
				"int8",
				"int16",
				"int32",
				"int64",
				"int",
				"float32",
				"float64",
			},
		},
	)
}

func NumValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NumValueT,
			TypeStrings: NodeTypes,
			Value: op.And{
				op.Optional(
					'-',
				),
				Digit,
				op.MinZero(
					op.And{
						op.Optional(
							'_',
						),
						Digit,
					},
				),
				op.Optional(
					op.And{
						'.',
						Digit,
						op.MinZero(
							op.And{
								op.Optional(
									'_',
								),
								Digit,
							},
						),
					},
				),
			},
		},
	)
}

func OptValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        OptValueT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"opt",
				Spp,
				op.Or{
					Num,
					Bool,
					Null,
					Text,
					Record,
					Variant,
					Principal,
					Vec,
					Blob,
				},
			},
		},
	)
}

func Principal(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        PrincipalT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"principal",
				Spp,
				TextValue,
			},
		},
	)
}

func Record(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        RecordT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"record",
				Sp,
				'{',
				Ws,
				op.Optional(
					RecordFields,
				),
				Ws,
				'}',
			},
		},
	)
}

func RecordField(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        RecordFieldT,
			TypeStrings: NodeTypes,
			Value: op.And{
				Id,
				Sp,
				'=',
				Sp,
				Value,
			},
		},
	)
}

func RecordFields(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			RecordField,
			Sp,
			op.MinZero(
				op.And{
					';',
					Ws,
					RecordField,
					Sp,
				},
			),
			op.Optional(
				';',
			),
		},
	)
}

func Sp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			' ',
		),
	)
}

func Spp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			' ',
			Sp,
		},
	)
}

func Text(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TextT,
			TypeStrings: NodeTypes,
			Value: op.And{
				TextValue,
				op.Optional(
					op.And{
						Sp,
						':',
						Sp,
						"text",
					},
				),
			},
		},
	)
}

func TextValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TextValueT,
			TypeStrings: NodeTypes,
			Value: op.And{
				'"',
				op.MinZero(
					Char,
				),
				'"',
			},
		},
	)
}

func Utf(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Ascii,
			UtfEnc,
		},
	)
}

func UtfEnc(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			op.And{
				parser.CheckRuneRange(0x00C2, 0x00DF),
				Utfcont,
			},
			op.And{
				0x00E0,
				parser.CheckRuneRange(0x00A0, 0x00BF),
				Utfcont,
			},
			op.And{
				0x00ED,
				parser.CheckRuneRange(0x0080, 0x009F),
				Utfcont,
			},
			op.And{
				parser.CheckRuneRange(0x00E1, 0x00EC),
				op.Repeat(2,
					Utfcont,
				),
			},
			op.And{
				parser.CheckRuneRange(0x00EE, 0x00EF),
				op.Repeat(2,
					Utfcont,
				),
			},
			op.And{
				0x00F0,
				parser.CheckRuneRange(0x0090, 0x00BF),
				op.Repeat(2,
					Utfcont,
				),
			},
			op.And{
				0x00F4,
				parser.CheckRuneRange(0x0080, 0x008F),
				op.Repeat(2,
					Utfcont,
				),
			},
			op.And{
				parser.CheckRuneRange(0x00F1, 0x00F3),
				op.Repeat(3,
					Utfcont,
				),
			},
		},
	)
}

func Utfcont(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(parser.CheckRuneRange(0x0080, 0x00BF))
}

func Value(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			OptValue,
			Num,
			Bool,
			Null,
			Text,
			Record,
			Variant,
			Principal,
			Vec,
			Blob,
		},
	)
}

func Values(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ValuesT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				op.And{
					'(',
					Sp,
					op.Optional(
						op.And{
							Value,
							op.MinZero(
								op.And{
									Sp,
									',',
									Sp,
									Value,
								},
							),
						},
					),
					Sp,
					')',
				},
				Value,
			},
		},
	)
}

func Variant(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        VariantT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"variant",
				Sp,
				'{',
				Ws,
				VariantField,
				Ws,
				'}',
			},
		},
	)
}

func VariantField(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			Id,
			op.Optional(
				op.And{
					Sp,
					'=',
					Sp,
					Value,
				},
			),
			op.Optional(
				';',
			),
		},
	)
}

func Vec(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        VecT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"vec",
				Sp,
				'{',
				Ws,
				op.Optional(
					VecFields,
				),
				Ws,
				'}',
			},
		},
	)
}

func VecFields(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			Value,
			Sp,
			op.MinZero(
				op.And{
					';',
					Ws,
					Value,
					Sp,
				},
			),
			op.Optional(
				';',
			),
		},
	)
}

func Ws(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			op.Or{
				' ',
				0x0009,
				0x000A,
				0x000D,
				op.And{
					0x000D,
					0x000A,
				},
			},
		),
	)
}
