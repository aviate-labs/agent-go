// Do not edit. This file is auto-generated.
// Grammar: CANDID-TEST (v0.1.0) github.com/di-wu/candid-go/internal/candidtest

package candidtest

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
)

// Node Types
const (
	Unknown = iota

	// CANDID-TEST (github.com/di-wu/candid-go/internal/candidtest)

	TestDataT    // 001
	CommentTextT // 002
	TestT        // 003
	TestGoodT    // 004
	TestBadT     // 005
	TestTestT    // 006
	NullT        // 007
	BoolT        // 008
	NatT         // 009
	IntT         // 010
	FloatT       // 011
	BaseT        // 012
	TextT        // 013
	ReservedT    // 014
	EmptyT       // 015
	OptT         // 016
	TextInputT   // 017
	BlobInputT   // 018
	DescriptionT // 019
)

// Token Definitions
const (
	// CANDID-TEST (github.com/di-wu/candid-go/internal/candidtest)

	ESC = 0x005C // \
)

var NodeTypes = []string{
	"UNKNOWN",

	// CANDID-TEST (github.com/di-wu/candid-go/internal/candidtest)

	"TestData",
	"CommentText",
	"Test",
	"TestGood",
	"TestBad",
	"TestTest",
	"Null",
	"Bool",
	"Nat",
	"Int",
	"Float",
	"Base",
	"Text",
	"Reserved",
	"Empty",
	"Opt",
	"TextInput",
	"BlobInput",
	"Description",
}

func Ascii(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		parser.CheckRuneRange(0x0020, 0x0021),
		parser.CheckRuneRange(0x0023, 0x005B),
		parser.CheckRuneRange(0x005D, 0x007E),
	})
}

func Base(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BaseT,
			TypeStrings: NodeTypes,
			Value: op.MinOne(
				Digit,
			),
		},
	)
}

func BlobInput(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BlobInputT,
			TypeStrings: NodeTypes,
			Value:       String,
		},
	)
}

func BlobInputTmpl(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			"blob \"",
			BlobInput,
			'"',
		},
	)
}

func Bool(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        BoolT,
			TypeStrings: NodeTypes,
			Value:       "bool",
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
		op.Or{
			op.And{
				"/*",
				Ws,
				MultiComment,
				Ws,
				"*/",
			},
			op.And{
				op.And{
					"//",
					op.Optional(
						CommentText,
					),
				},
				EndLine,
			},
		},
	)
}

func CommentText(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        CommentTextT,
			TypeStrings: NodeTypes,
			Value: op.MinZero(
				op.And{
					op.Not{
						EndLine,
					},
					parser.CheckRuneRange(0x0000, 0x0010FFFF),
				},
			),
		},
	)
}

func Description(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        DescriptionT,
			TypeStrings: NodeTypes,
			Value: op.And{
				'"',
				String,
				'"',
			},
		},
	)
}

func Digit(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(parser.CheckRuneRange('0', '9'))
}

func Empty(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        EmptyT,
			TypeStrings: NodeTypes,
			Value:       "empty",
		},
	)
}

func EndLine(p *ast.Parser) (*ast.Node, error) {
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

func Float(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        FloatT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"float",
				Base,
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

func Input(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			BlobInputTmpl,
			TextInputTmpl,
		},
	)
}

func Int(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        IntT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"int",
				op.Optional(
					Base,
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

func MultiComment(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			op.And{
				op.Not{
					"*/",
				},
				op.Optional(
					CommentText,
				),
				EndLine,
			},
		),
	)
}

func Nat(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        NatT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"nat",
				op.Optional(
					Base,
				),
			},
		},
	)
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

func Opt(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        OptT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"opt ",
				Values,
			},
		},
	)
}

func Reserved(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        ReservedT,
			TypeStrings: NodeTypes,
			Value:       "reserved",
		},
	)
}

func String(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			Char,
		),
	)
}

func Test(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TestT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"assert ",
				Input,
				Ws,
				op.Or{
					TestGoodTmpl,
					TestBadTmpl,
					TestTest,
				},
				op.Optional(
					op.And{
						' ',
						Description,
					},
				),
				';',
			},
		},
	)
}

func TestBad(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TestBadT,
			TypeStrings: NodeTypes,
			Value:       ValuesBr,
		},
	)
}

func TestBadTmpl(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			"!:",
			Ws,
			TestBad,
		},
	)
}

func TestData(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TestDataT,
			TypeStrings: NodeTypes,
			Value: op.MinOne(
				op.Or{
					Comment,
					Test,
					EndLine,
				},
			),
		},
	)
}

func TestGood(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TestGoodT,
			TypeStrings: NodeTypes,
			Value:       ValuesBr,
		},
	)
}

func TestGoodTmpl(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			':',
			Ws,
			TestGood,
		},
	)
}

func TestTest(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TestTestT,
			TypeStrings: NodeTypes,
			Value: op.And{
				"==",
				Ws,
				Input,
				Ws,
				':',
				Ws,
				ValuesBr,
			},
		},
	)
}

func Text(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TextT,
			TypeStrings: NodeTypes,
			Value:       "text",
		},
	)
}

func TextInput(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        TextInputT,
			TypeStrings: NodeTypes,
			Value:       String,
		},
	)
}

func TextInputTmpl(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.And{
			'"',
			TextInput,
			'"',
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

func Values(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Null,
			Bool,
			Nat,
			Int,
			Float,
			Text,
			Reserved,
			Empty,
			Opt,
		},
	)
}

func ValuesBr(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			"()",
			op.And{
				'(',
				Values,
				op.MinZero(
					op.And{
						", ",
						Values,
					},
				),
				')',
			},
		},
	)
}

func Ws(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.MinZero(
			op.Or{
				' ',
				0x0009,
				EndLine,
			},
		),
	)
}
