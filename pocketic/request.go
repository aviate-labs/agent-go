package pocketic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"io"
	"net/http"
)

var headers = http.Header{
	"content-type":          []string{"application/json"},
	"processing-timeout-ms": []string{"300000"},
}

func checkResponse(resp *http.Response, statusCode int, v any) error {
	if resp.StatusCode != statusCode {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if v == nil {
		// No need to decode the response body.
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
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
	req.Header = headers
	return req, nil
}

// AwaitCall awaits an update call submitted previously by `submit_call_with_effective_principal`.
func (pic PocketIC) AwaitCall(messageID RawMessageID) ([]byte, error) {
	var resp Result[WASMResult[[]byte]]
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/await_ingress_message", pic.instanceURL()),
		http.StatusOK,
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

// QueryCall executes a query call on a canister.
func (pic PocketIC) QueryCall(canisterID principal.Principal, sender principal.Principal, method string, payload []byte) ([]byte, error) {
	raw, err := pic.canisterCall("read/query", canisterID, new(RawEffectivePrincipalNone), sender, method, payload)
	if err != nil {
		return nil, err
	}
	return *raw, nil
}

// SubmitCall submits an update call (without executing it immediately).
func (pic PocketIC) SubmitCall(
	canisterID principal.Principal,
	sender principal.Principal,
	method string,
	payload []byte,
) (*RawMessageID, error) {
	return pic.SubmitCallWithEP(
		canisterID,
		new(RawEffectivePrincipalNone),
		sender,
		method,
		payload,
	)
}

// SubmitCallWithEP submits an update call with a provided effective principal (without executing it immediately).
func (pic PocketIC) SubmitCallWithEP(
	canisterID principal.Principal,
	effectivePrincipal RawEffectivePrincipal,
	sender principal.Principal,
	method string,
	payload []byte,
) (*RawMessageID, error) {
	var resp RawSubmitIngressResult
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/submit_ingress_message", pic.instanceURL()),
		http.StatusOK,
		RawCanisterCall{
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
func (pic PocketIC) UpdateCall(canisterID principal.Principal, sender principal.Principal, method string, payload []byte) ([]byte, error) {
	return pic.updateCallWithEP(canisterID, &RawEffectivePrincipalCanisterID{CanisterID: canisterID.Raw}, sender, method, payload)
}

// canisterCall calls the canister endpoint with the provided arguments.
func (pic PocketIC) canisterCall(endpoint string, canisterID principal.Principal, effectivePrincipal RawEffectivePrincipal, sender principal.Principal, method string, payload []byte) (*Base64EncodedBlob, error) {
	var resp RawCanisterResult
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/%s", pic.instanceURL(), endpoint),
		http.StatusOK,
		RawCanisterCall{
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
	return resp.Ok.Reply, nil
}

func (pic PocketIC) do(method, url string, statusCode int, input, output any) error {
	pic.logger.Printf("[POCKETIC] %s %s %+v", method, url, input)
	req, err := newRequest(method, url, input)
	if err != nil {
		return err
	}
	resp, err := pic.client.Do(req)
	if err != nil {
		return err
	}
	return checkResponse(resp, statusCode, output)
}

// updateCallWithEP calls SubmitCallWithEP and AwaitCall in sequence.
func (pic PocketIC) updateCallWithEP(canisterID principal.Principal, effectivePrincipal RawEffectivePrincipal, sender principal.Principal, method string, payload []byte) ([]byte, error) {
	messageID, err := pic.SubmitCallWithEP(canisterID, effectivePrincipal, sender, method, payload)
	if err != nil {
		return nil, err
	}
	return pic.AwaitCall(*messageID)
}
