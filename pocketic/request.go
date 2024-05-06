package pocketic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"io"
	"net/http"
)

var headers = func() http.Header {
	return http.Header{
		"content-type":          []string{"application/json"},
		"processing-timeout-ms": []string{"300000"},
	}
}

func newRequest(method, url string, body any) (*http.Request, error) {
	var bodyBytes io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyBytes = bytes.NewBuffer(raw)
	}
	req, err := http.NewRequest(method, url, bodyBytes)
	if err != nil {
		return nil, err
	}
	req.Header = headers()
	return req, nil
}

// AwaitCall awaits an update call submitted previously by `submit_call_with_effective_principal`.
func (pic PocketIC) AwaitCall(messageID MessageID) ([]byte, error) {
	var resp result[wasmResult[[]byte]]
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/await_ingress_message", pic.InstanceURL()),
		messageID,
		&resp,
	); err != nil {
		return nil, err
	}
	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.Ok.Reject != nil {
		return nil, resp.Ok.Reject
	}
	return *resp.Ok.Reply, nil
}

// ExecuteCall executes an update call on a canister.
func (pic PocketIC) ExecuteCall(
	canisterID principal.Principal,
	effectivePrincipal EffectivePrincipal,
	sender principal.Principal,
	method string,
	payload []byte,
) ([]byte, error) {
	var resp rawCanisterResult
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/execute_ingress_message", pic.InstanceURL()),
		rawCanisterCall{
			CanisterID:         canisterID.Raw,
			EffectivePrincipal: effectivePrincipal,
			Method:             method,
			Payload:            payload,
			Sender:             sender.Raw,
		},
		&resp,
	); err != nil {
		return nil, err
	}
	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.Ok.Reject != nil {
		return nil, resp.Ok.Reject
	}
	return *resp.Ok.Reply, nil
}

// QueryCall executes a query call on a canister.
func (pic PocketIC) QueryCall(canisterID principal.Principal, sender principal.Principal, method string, args []any, ret []any) error {
	payload, err := idl.Marshal(args)
	if err != nil {
		return err
	}
	raw, err := pic.canisterCall("read/query", canisterID, new(EffectivePrincipalNone), sender, method, payload)
	if err != nil {
		return err
	}
	if err := idl.Unmarshal(raw, ret); err != nil {
		return err
	}
	return nil
}

// SubmitCall submits an update call (without executing it immediately).
func (pic PocketIC) SubmitCall(
	canisterID principal.Principal,
	sender principal.Principal,
	method string,
	payload []byte,
) (*MessageID, error) {
	return pic.SubmitCallWithEP(
		canisterID,
		new(EffectivePrincipalNone),
		sender,
		method,
		payload,
	)
}

// SubmitCallWithEP submits an update call with a provided effective principal (without executing it immediately).
func (pic PocketIC) SubmitCallWithEP(
	canisterID principal.Principal,
	effectivePrincipal EffectivePrincipal,
	sender principal.Principal,
	method string,
	payload []byte,
) (*MessageID, error) {
	var resp rawSubmitIngressResult
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/submit_ingress_message", pic.InstanceURL()),
		rawCanisterCall{
			CanisterID:         canisterID.Raw,
			EffectivePrincipal: effectivePrincipal,
			Method:             method,
			Payload:            payload,
			Sender:             sender.Raw,
		},
		&resp,
	); err != nil {
		return nil, err
	}
	if resp.Err != nil {
		return nil, resp.Err
	}
	return resp.Ok, nil
}

// UpdateCall executes an update call on a canister.
func (pic PocketIC) UpdateCall(canisterID principal.Principal, sender principal.Principal, method string, args []any, ret []any) error {
	payload, err := idl.Marshal(args)
	if err != nil {
		return err
	}
	raw, err := pic.updateCallWithEP(canisterID, &EffectivePrincipalCanisterID{CanisterID: canisterID.Raw}, sender, method, payload)
	if err != nil {
		return err
	}
	return idl.Unmarshal(raw, ret)
}

// canisterCall calls the canister endpoint with the provided arguments.
func (pic PocketIC) canisterCall(endpoint string, canisterID principal.Principal, effectivePrincipal EffectivePrincipal, sender principal.Principal, method string, payload []byte) ([]byte, error) {
	var resp rawCanisterResult
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/%s", pic.InstanceURL(), endpoint),
		rawCanisterCall{
			CanisterID:         canisterID.Raw,
			EffectivePrincipal: effectivePrincipal,
			Method:             method,
			Payload:            payload,
			Sender:             sender.Raw,
		},
		&resp,
	); err != nil {
		return nil, err
	}
	if resp.Err != nil {
		return nil, resp.Err
	}
	if resp.Ok.Reject != nil {
		return nil, resp.Ok.Reject
	}
	return *resp.Ok.Reply, nil
}

// updateCallWithEP calls SubmitCallWithEP and AwaitCall in sequence.
func (pic PocketIC) updateCallWithEP(canisterID principal.Principal, effectivePrincipal EffectivePrincipal, sender principal.Principal, method string, payload []byte) ([]byte, error) {
	messageID, err := pic.SubmitCallWithEP(canisterID, effectivePrincipal, sender, method, payload)
	if err != nil {
		return nil, err
	}
	return pic.AwaitCall(*messageID)
}
