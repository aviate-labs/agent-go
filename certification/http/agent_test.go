package http_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certification/http"
	"github.com/aviate-labs/agent-go/certification/http/certexp"
	"github.com/aviate-labs/agent-go/principal"
)

func TestAgent_HttpRequest(t *testing.T) {
	canisterId, _ := principal.Decode("rdmx6-jaaaa-aaaaa-aaadq-cai")
	a, err := http.NewAgent(canisterId, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	certVersion := uint16(2)
	if _, err := a.HttpRequest(http.Request{
		Method:             "GET",
		Url:                "/",
		CertificateVersion: &certVersion,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCalculateRequestHash(t *testing.T) {
	req := func(url string) *http.Request {
		return &http.Request{
			Method: "POST",
			Url:    url,
			Headers: []http.HeaderField{
				{"Accept-Language", "en"},
				{"Accept-Language", "en-US"},
				{"Host", "https://ic0.app"},
			},
			Body: []byte{0, 1, 2, 3, 4, 5, 6},
		}
	}

	t.Run("without query", func(t *testing.T) {
		hash, err := http.CalculateRequestHash(
			req("https://ic0.app"),
			&certexp.CertificateExpressionRequestCertification{
				CertifiedRequestHeaders: []string{"host"},
			},
		)
		if err != nil {
			t.Error(err)
		}
		if h := fmt.Sprintf("%x", hash); h != "10796453466efb3e333891136b8a5931269f77e40ead9d437fcee94a02fa833c" {
			t.Error("hash mismatch", h)
		}
	})

	t.Run("with query", func(t *testing.T) {
		hash, err := http.CalculateRequestHash(
			req("https://ic0.app?q=hello+world&name=foo&name=bar&color=purple"),
			&certexp.CertificateExpressionRequestCertification{
				CertifiedRequestHeaders:  []string{"host"},
				CertifiedQueryParameters: []string{"q", "name"},
			},
		)
		if err != nil {
			t.Error(err)
		}
		if h := fmt.Sprintf("%x", hash); h != "3ade1c9054f05bc8bcebd3fd7b884078a6e67c63e5ac4a639fa46a47f5a955c9" {
			t.Error("hash mismatch", h)
		}
	})
}

func TestResponse_Verify_V1(t *testing.T) {
	canisterId, _ := principal.Decode("rdmx6-jaaaa-aaaaa-aaadq-cai")
	a, err := http.NewAgent(canisterId, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}

	a.EnableLegacyMode() // Enable legacy mode.
	path := "/"
	certVersion := uint16(1)
	req := http.Request{
		Method:             "GET",
		Url:                path,
		CertificateVersion: &certVersion,
	}
	resp, err := a.HttpRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.VerifyResponse(path, &req, resp); err != nil {
		t.Error(err)
	}
}

func TestResponse_Verify_V2(t *testing.T) {
	canisterId, _ := principal.Decode("rdmx6-jaaaa-aaaaa-aaadq-cai")
	a, err := http.NewAgent(canisterId, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	path := "/index.html"
	certVersion := uint16(2)
	req := http.Request{
		Method:             "GET",
		Url:                path,
		CertificateVersion: &certVersion,
	}
	resp, err := a.HttpRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.VerifyResponse(path, &req, resp); err != nil {
		t.Error(err)
	}
}
