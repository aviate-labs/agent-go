package agent

import (
	"context"

	"github.com/fxamacker/cbor/v2"
	"github.com/niccolofant/agent-go/certification"
	"github.com/niccolofant/agent-go/certification/hashtree"
	"github.com/niccolofant/agent-go/principal"
	"google.golang.org/protobuf/proto"
)

// CallAndWait calls a method on a canister and waits for the result.
func (c APIRequest[_, Out]) CallAndWait(out Out) error {
	return c.CallAndWaitWithContext(c.a.ctx, out)
}

// CallAndWaitWithContext is like CallAndWait but uses the given context as the parent
// of the per-request timeouts and the polling loop, letting the caller cancel an
// in-flight update call.
func (c APIRequest[_, Out]) CallAndWaitWithContext(ctx context.Context, out Out) error {
	c.a.logger.Printf("[AGENT] CALL %s %s (%x)", c.effectiveCanisterID, c.methodName, c.requestID)
	rawCertificate, err := c.a.call(ctx, c.effectiveCanisterID, c.data)
	if err != nil {
		return err
	}
	if len(rawCertificate) != 0 {
		var certificate certification.Certificate
		if err := cbor.Unmarshal(rawCertificate, &certificate); err != nil {
			return err
		}
		path := []hashtree.Label{hashtree.Label("request_status"), c.requestID[:]}
		if raw, err := certificate.Tree.Lookup(append(path, hashtree.Label("reply"))...); err == nil {
			return c.unmarshal(raw, out)
		}

		rejectCode, err := certificate.Tree.Lookup(append(path, hashtree.Label("reject_code"))...)
		if err != nil {
			return err
		}
		message, err := certificate.Tree.Lookup(append(path, hashtree.Label("reject_message"))...)
		if err != nil {
			return err
		}
		errorCode, err := certificate.Tree.Lookup(append(path, hashtree.Label("error_code"))...)
		if err != nil {
			return err
		}
		return preprocessingError{
			RejectCode: uint64FromBytes(rejectCode),
			Message:    string(message),
			ErrorCode:  string(errorCode),
		}
	}

	raw, err := c.a.poll(ctx, c.effectiveCanisterID, c.requestID)
	if err != nil {
		return err
	}
	return c.unmarshal(raw, out)
}

// Call calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Call(canisterID principal.Principal, methodName string, in []any, out []any) error {
	call, err := a.CreateCandidAPIRequest(RequestTypeCall, canisterID, methodName, in...)
	if err != nil {
		return err
	}
	return call.CallAndWait(out)
}

// CallProto calls a method on a canister and unmarshals the result into the given proto message.
func (a Agent) CallProto(canisterID principal.Principal, methodName string, in, out proto.Message) error {
	call, err := a.CreateProtoAPIRequest(RequestTypeCall, canisterID, methodName, in)
	if err != nil {
		return err
	}
	return call.CallAndWait(out)
}

// CallRaw submits an update call with an opaque argument and returns the raw reply bytes.
// Neither the argument nor the reply is interpreted.
//
// Example:
//
//	reply, err := a.CallRaw(canisterID, "ingest", cborBytes)
func (a Agent) CallRaw(canisterID principal.Principal, methodName string, arg []byte) ([]byte, error) {
	call, err := a.CreateRawAPIRequest(RequestTypeCall, canisterID, methodName, arg)
	if err != nil {
		return nil, err
	}
	var out []byte
	if err := call.CallAndWait(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// CallWithContext is like Call but uses the given context as the parent of the
// per-request timeouts and the polling loop, letting the caller cancel an in-flight
// update call.
func (a Agent) CallWithContext(ctx context.Context, canisterID principal.Principal, methodName string, in []any, out []any) error {
	call, err := a.CreateCandidAPIRequest(RequestTypeCall, canisterID, methodName, in...)
	if err != nil {
		return err
	}
	return call.CallAndWaitWithContext(ctx, out)
}

// CallWithEffectiveCanisterID is like Call but lets the caller supply the effective
// canister ID. Needed for management-canister methods whose args carry no canister_id
// (create_canister, provisional_create_canister_with_cycles).
func (a Agent) CallWithEffectiveCanisterID(canisterID, effectiveCanisterID principal.Principal, methodName string, in, out []any) error {
	call, err := a.CreateCandidAPIRequest(RequestTypeCall, canisterID, methodName, in...)
	if err != nil {
		return err
	}
	return call.WithEffectiveCanisterID(effectiveCanisterID).CallAndWait(out)
}
