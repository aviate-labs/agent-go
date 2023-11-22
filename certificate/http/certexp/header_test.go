package certexp

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"testing"
)

func Test_certificateExpressionHeader(t *testing.T) {
	header := "default_certification(ValidationArgs{certification:Certification{no_request_certification: Empty{},response_certification:ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}}})"
	p, err := ast.New([]byte(header))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Expect(op.And{Value, parser.EOD}); err != nil {
		t.Fatal(err)
	}
}

func Test_certificateStringList(t *testing.T) {
	for _, test := range []string{
		"[]", `[""]`, `["a"]`, `["a" "b"]`, `["a" "b" "c"]`,
	} {
		p, err := ast.New([]byte(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Expect(op.And{StringList, parser.EOD}); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_certification(t *testing.T) {
	for _, test := range []string{
		"Certification{no_request_certification: Empty{},response_certification:ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}}",
		"Certification{ no_request_certification: Empty{}, response_certification: ResponseCertification{ response_header_exclusions: ResponseHeaderList{ headers: [] } } }",
	} {
		p, err := ast.New([]byte(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Expect(op.And{Certification, parser.EOD}); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_requestCertification(t *testing.T) {
	for _, test := range []string{
		"RequestCertification{certified_request_headers:[],certified_query_parameters:[]}",
		"RequestCertification{ certified_request_headers: [], certified_query_parameters: [] }",
	} {
		p, err := ast.New([]byte(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Expect(op.And{RequestCertification, parser.EOD}); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_responseCertification(t *testing.T) {
	for _, test := range []string{
		"ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}",
		"ResponseCertification{ response_header_exclusions: ResponseHeaderList{ headers: [] } }",
		"ResponseCertification{certified_response_headers:ResponseHeaderList{headers:[]}}",
	} {
		p, err := ast.New([]byte(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Expect(op.And{ResponseCertification, parser.EOD}); err != nil {
			t.Fatal(err)
		}
	}
}
