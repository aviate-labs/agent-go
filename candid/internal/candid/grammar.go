// Do not edit. This file is auto-generated.
// Grammar: CANDID (v0.1.1) github.com/di-wu/candid-go/internal/candid

package candid

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
)

// Node Types
const (
	Unknown = iota

	// CANDID (github.com/di-wu/candid-go/internal/candid)

	ProgT        // 001
	TypeT        // 002
	ImportT      // 003
	ActorT       // 004
	ActorTypeT   // 005
	MethTypeT    // 006
	FuncTypeT    // 007
	FuncAnnT     // 008
	TupTypeT     // 009
	ArgTypeT     // 010
	FieldTypeT   // 011
	DataTypeT    // 012
	PrimTypeT    // 013
	BlobT        // 014
	OptT         // 015
	VecT         // 016
	RecordT      // 017
	VariantT     // 018
	FuncT        // 019
	ServiceT     // 020
	PrincipalT   // 021
	IdT          // 022
	TextT        // 023
	NatT         // 024
	CommentTextT // 025
)

// Token Definitions
const (
	// CANDID (github.com/di-wu/candid-go/internal/candid)

	ESC = 0x005C // \
)

var NodeTypes = []string{
	"UNKNOWN",

	// CANDID (github.com/di-wu/candid-go/internal/candid)

	"Prog",
	"Type",
	"Import",
	"Actor",
	"ActorType",
	"MethType",
	"FuncType",
	"FuncAnn",
	"TupType",
	"ArgType",
	"FieldType",
	"DataType",
	"PrimType",
	"Blob",
	"Opt",
	"Vec",
	"Record",
	"Variant",
	"Func",
	"Service",
	"Principal",
	"Id",
	"Text",
	"Nat",
	"CommentText",
}

func Actor(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ActorT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"service",
				op.Optional(
					Sp,
				),
				op.Optional(
					op.And{
						Id,
						Sp,
					},
				),
				':',
				Sp,
				op.Optional(
					op.And{
						TupType,
						Sp,
						"->",
						Ws,
					},
				),
				op.Or{
					ActorType,
					Id,
				},
			},
		},
	)
}

func ActorType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ActorTypeT,
			TypeStrings: NodeTypes,
			Value: op.And{
				'{',
				op.Optional(
					op.And{
						Ws,
						MethType,
						op.MinZero(
							op.And{
								';',
								Ws,
								MethType,
							},
						),
						op.Optional(
							';',
						),
						Ws,
					},
				),
				'}',
			},
		},
	)
}

func ArgType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ArgTypeT,
			TypeStrings: NodeTypes,
			Value: op.And{
				op.Optional(
					op.And{
						Name,
						op.Optional(
							Sp,
						),
						':',
						Sp,
					},
				),
				DataType,
			},
		},
	)
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
			Value:       "blob",
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

func Comment(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			"//",
			CommentText,
			Nl,
		},
	)
}

func CommentText(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        CommentTextT,
			TypeStrings: NodeTypes,
			Value: op.MinZero(
				op.Or{
					Ascii,
					0x0022,
					0x0027,
					0x0060,
				},
			),
		},
	)
}

func ConsType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Blob,
			Opt,
			Vec,
			Record,
			Variant,
		},
	)
}

func DataType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        DataTypeT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				ConsType,
				RefType,
				PrimType,
				Id,
			},
		},
	)
}

func Def(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Type,
			Import,
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

func FieldType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        FieldTypeT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				op.And{
					op.Optional(
						op.And{
							op.Or{
								Nat,
								Name,
							},
							op.Optional(
								Sp,
							),
							':',
							Sp,
						},
					),
					DataType,
				},
				Nat,
				Name,
			},
		},
	)
}

func Fields(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			FieldType,
			op.MinZero(
				op.And{
					';',
					Ws,
					FieldType,
				},
			),
			op.Optional(
				';',
			),
		},
	)
}

func Func(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        FuncT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"func",
				Sp,
				FuncType,
			},
		},
	)
}

func FuncAnn(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        FuncAnnT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				"oneway",
				"query",
			},
		},
	)
}

func FuncType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        FuncTypeT,
			TypeStrings: NodeTypes,
			Value: op.And{
				TupType,
				op.Optional(
					op.And{
						Sp,
						"->",
						Ws,
						TupType,
						op.Optional(
							op.And{
								Sp,
								FuncAnn,
							},
						),
					},
				),
			},
		},
	)
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

func Import(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ImportT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"import",
				Sp,
				Text,
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

func MethType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        MethTypeT,
			TypeStrings: NodeTypes,
			Value: op.And{
				Name,
				op.Optional(
					Sp,
				),
				':',
				Ws,
				op.Or{
					FuncType,
					Id,
				},
			},
		},
	)
}

func Name(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Id,
			Text,
		},
	)
}

func Nat(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NatT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				op.And{
					"0x",
					HexNum,
				},
				Num,
			},
		},
	)
}

func Nl(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			0x000A,
			0x000D,
			op.And{
				0x000D,
				0x000A,
			},
		},
	)
}

func Num(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
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
	)
}

func NumType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
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
	)
}

func Opt(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        OptT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"opt",
				Sp,
				DataType,
			},
		},
	)
}

func PrimType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        PrimTypeT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				NumType,
				"bool",
				"text",
				"null",
				"reserved",
				"empty",
			},
		},
	)
}

func Principal(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        PrincipalT,
			TypeStrings: NodeTypes,
			Value:       "principal",
		},
	)
}

func Prog(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ProgT,
			TypeStrings: NodeTypes,
			Value: op.And{
				op.Optional(
					op.And{
						Ws,
						Def,
						op.MinZero(
							op.And{
								';',
								Ws,
								Def,
							},
						),
					},
				),
				op.Optional(
					';',
				),
				Ws,
				op.Optional(
					op.And{
						Ws,
						Actor,
						op.MinZero(
							op.And{
								';',
								Ws,
								Actor,
							},
						),
					},
				),
				op.Optional(
					';',
				),
				Ws,
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
					Fields,
				),
				Ws,
				'}',
			},
		},
	)
}

func RefType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Func,
			Service,
			Principal,
		},
	)
}

func Service(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ServiceT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"service",
				Sp,
				ActorType,
			},
		},
	)
}

func Sp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinOne(
			' ',
		),
	)
}

func Text(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TextT,
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

func TupType(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TupTypeT,
			TypeStrings: NodeTypes,
			Value: op.Or{
				op.And{
					'(',
					op.Optional(
						op.And{
							ArgType,
							op.MinZero(
								op.And{
									',',
									Sp,
									ArgType,
								},
							),
						},
					),
					op.Optional(
						Sp,
					),
					')',
				},
				ArgType,
			},
		},
	)
}

func Type(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TypeT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"type",
				Sp,
				Id,
				Sp,
				'=',
				Sp,
				DataType,
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
				op.Optional(
					Fields,
				),
				Ws,
				'}',
			},
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
				DataType,
			},
		},
	)
}

func Ws(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			op.Or{
				Sp,
				0x0009,
				Comment,
				Nl,
			},
		),
	)
}
