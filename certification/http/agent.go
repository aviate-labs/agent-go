package http

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"strings"
)

type Agent struct {
	canisterId          principal.Principal
	supportsV1          bool
	supportsV2, forceV2 bool
	*agent.Agent
}

func NewAgent(canisterId principal.Principal, cfg agent.Config) (*Agent, error) {
	a, err := agent.New(cfg)
	if err != nil {
		return nil, err
	}

	var supportsV1, supportsV2 bool
	if raw, err := a.GetCanisterMetadata(canisterId, "supported_certificate_versions"); err == nil {
		for v := range strings.SplitSeq(string(raw), ",") {
			switch v {
			case "1":
				supportsV1 = true
			case "2":
				supportsV2 = true
			}
		}
	}

	return &Agent{
		canisterId: canisterId,
		supportsV1: supportsV1,
		supportsV2: supportsV2,
		forceV2:    true,
		Agent:      a,
	}, nil
}

func (a *Agent) EnableLegacyMode() {
	a.forceV2 = false
}

func (a Agent) HttpRequest(request Request) (*Response, error) {
	var r0 Response
	if err := a.Query(
		a.canisterId,
		"http_request",
		[]any{request},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

func (a Agent) HttpRequestStreamingCallback(token StreamingCallbackToken) (**StreamingCallbackHttpResponse, error) {
	var r0 *StreamingCallbackHttpResponse
	if err := a.Query(
		a.canisterId,
		"http_request_streaming_callback",
		[]any{token},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

type HeaderField struct {
	Field0 string `ic:"0"`
	Field1 string `ic:"1"`
}

type Key = string

type Request struct {
	Method             string        `ic:"method"`
	Url                string        `ic:"url"`
	Headers            []HeaderField `ic:"headers"`
	Body               []byte        `ic:"body"`
	CertificateVersion *uint16       `ic:"certificate_version,omitempty"`
}

type Response struct {
	StatusCode        uint16             `ic:"status_code"`
	Headers           []HeaderField      `ic:"headers"`
	Body              []byte             `ic:"body"`
	Upgrade           *bool              `ic:"upgrade,omitempty"`
	StreamingStrategy *StreamingStrategy `ic:"streaming_strategy,omitempty"`
}

type StreamingCallbackHttpResponse struct {
	Body  []byte                  `ic:"body"`
	Token *StreamingCallbackToken `ic:"token,omitempty"`
}

type StreamingCallbackToken struct {
	Key             Key     `ic:"key"`
	ContentEncoding string  `ic:"content_encoding"`
	Index           idl.Nat `ic:"index"`
	Sha256          *[]byte `ic:"sha256,omitempty"`
}

type StreamingStrategy struct {
	Callback *struct {
		Callback struct { /* NOT SUPPORTED */
		} `ic:"callback"`
		Token StreamingCallbackToken `ic:"token"`
	} `ic:"Callback,variant"`
}
