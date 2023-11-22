package http_test

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certificate/http"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
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
