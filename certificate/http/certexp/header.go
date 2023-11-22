package certexp

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"unicode/utf8"
)

const (
	Unknown = iota

	StringT
	ListT

	NoCertificationT
	CertificationT
	NoRequestCertificationT
	RequestCertificationT
	ResponseCertificationT
	ResponseHeaderListT
)

var (
	WS = op.MinZero(' ')

	Value          = op.And{"default_certification(", WS, ValidationArgs, WS, ")"}
	ValidationArgs = op.And{
		"ValidationArgs{",
		WS,
		op.Or{
			ast.Capture{
				Type:        NoCertificationT,
				TypeStrings: NodeTypes,
				Value:       op.And{"no_certification:", WS, "Empty{}"},
			},
			ast.Capture{
				Type:        CertificationT,
				TypeStrings: NodeTypes,
				Value:       op.And{"certification:", Certification}},
		},
		WS,
		"}",
	}
	Certification = op.And{
		"Certification{",
		WS,
		op.Or{
			ast.Capture{
				Type:        NoRequestCertificationT,
				TypeStrings: NodeTypes,
				Value:       op.And{"no_request_certification:", WS, "Empty{}"},
			},
			op.And{"request_certification:", RequestCertification},
		},
		WS,
		",",
		WS,
		"response_certification:",
		WS,
		ResponseCertification,
		WS,
		"}",
	}
	RequestCertification = ast.Capture{
		Type:        RequestCertificationT,
		TypeStrings: NodeTypes,
		Value: op.And{
			"RequestCertification{",
			WS,
			"certified_request_headers:", WS,
			StringList,
			WS, ",", WS,
			"certified_query_parameters:", WS,
			StringList,
			WS,
			"}",
		},
	}
	ResponseCertification = op.And{
		"ResponseCertification{",
		WS,
		ast.Capture{
			Type:        ResponseCertificationT,
			TypeStrings: NodeTypes,
			Value:       op.Or{"response_header_exclusions:", "certified_response_headers:"},
		},
		WS,
		ResponseHeaderList,
		WS,
		"}",
	}
	ResponseHeaderList = ast.Capture{
		Type:        ResponseHeaderListT,
		TypeStrings: NodeTypes,
		Value:       op.And{"ResponseHeaderList{", WS, "headers:", WS, StringList, WS, "}"},
	}

	Char   = op.And{op.Not{Value: op.Or{'0', '\n', '"'}}, parser.CheckRuneRange(0x00, utf8.MaxRune)}
	String = ast.Capture{
		Type:        StringT,
		TypeStrings: NodeTypes,
		Value:       op.And{'"', op.MinZero(Char), '"'},
	}
	StringList = ast.Capture{
		Type:        ListT,
		TypeStrings: NodeTypes,
		Value:       op.And{'[', WS, op.MinZero(op.And{String, WS}), ']'},
	}
)

var NodeTypes = []string{
	"UNKNOWN",

	"String",
	"List",

	"NoCertification",
	"Certification",
	"NoRequestCertification",
	"RequestCertification",
	"ResponseCertification",
	"ResponseHeaderList",
}
