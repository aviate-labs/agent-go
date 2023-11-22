package certexp

import (
	"fmt"
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
)

func parseStringList(n *ast.Node) []string {
	if n.Type != ListT {
		return nil
	}
	var s []string
	for _, n := range n.Children() {
		if n.Type != StringT {
			continue
		}
		s = append(s, n.Value)
	}
	return s
}

type CertificateExpression struct {
	Certification *CertificateExpressionCertification
}

func ParseCertificateExpression(expression string) (*CertificateExpression, error) {
	p, err := ast.New([]byte(expression))
	if err != nil {
		return nil, err
	}
	n, err := p.Expect(op.And{Value, parser.EOD})
	if err != nil {
		return nil, err
	}
	n = n.Children()[0]
	if n.Type == NoCertificationT {
		return &CertificateExpression{}, nil
	}
	exp := CertificateExpression{
		Certification: &CertificateExpressionCertification{},
	}
	if n := n.Children()[0]; n.Type == RequestCertificationT {
		exp.Certification.RequestCertification = &CertificateExpressionRequestCertification{
			CertifiedRequestHeaders:  parseStringList(n.Children()[0]),
			CertifiedQueryParameters: parseStringList(n.Children()[1]),
		}
	}
	switch t := n.Children()[1].Value; t {
	case "response_header_exclusions:":
		exp.Certification.ResponseCertification.ResponseHeaderExclusions = parseStringList(n.Children()[2])
	case "certified_response_headers:":
		exp.Certification.ResponseCertification.CertifiedResponseHeaders = parseStringList(n.Children()[2])
	default:
		return nil, fmt.Errorf("unknown response certification type: %s", t)
	}
	return &exp, nil
}

type CertificateExpressionCertification struct {
	RequestCertification  *CertificateExpressionRequestCertification
	ResponseCertification CertificateExpressionResponseCertification
}

type CertificateExpressionRequestCertification struct {
	CertifiedRequestHeaders  []string
	CertifiedQueryParameters []string
}

type CertificateExpressionResponseCertification struct {
	CertifiedResponseHeaders []string
	ResponseHeaderExclusions []string
}
