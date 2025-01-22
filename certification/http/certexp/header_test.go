package certexp

import (
	"testing"

	"github.com/0x51-dev/upeg/parser"
)

func Test_certificateExpressionHeader(t *testing.T) {
	header := "default_certification(ValidationArgs{certification:Certification{no_request_certification: Empty{},response_certification:ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}}})"
	p, err := parser.New([]rune(header))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(Value); err != nil {
		t.Fatal(err)
	}
}

func Test_certificateStringList(t *testing.T) {
	for _, test := range []string{
		"[]", `[""]`, `["a"]`, `["a" "b"]`, `["a" "b" "c"]`,
	} {
		p, err := parser.New([]rune(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ParseEOF(StringList); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_certification(t *testing.T) {
	for _, test := range []string{
		"Certification{no_request_certification: Empty{},response_certification:ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}}",
		"Certification{ no_request_certification: Empty{}, response_certification: ResponseCertification{ response_header_exclusions: ResponseHeaderList{ headers: [] } } }",
	} {
		p, err := parser.New([]rune(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ParseEOF(Certification); err != nil {
			t.Fatal(err)
		}
	}
}

func Test_requestCertification(t *testing.T) {
	for _, test := range []string{
		"RequestCertification{certified_request_headers:[],certified_query_parameters:[]}",
		"RequestCertification{ certified_request_headers: [], certified_query_parameters: [] }",
	} {
		p, err := parser.New([]rune(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ParseEOF(RequestCertification); err != nil {
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
		p, err := parser.New([]rune(test))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ParseEOF(ResponseCertification); err != nil {
			t.Fatal(err)
		}
	}
}
