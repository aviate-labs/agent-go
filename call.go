package agent

import (
	"fmt"

	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/protobuf/proto"
)

// CallAndWait calls a method on a canister and waits for the result.
func (c APIRequest[_, Out]) CallAndWait(out Out) error {
	c.a.logger.Printf("[AGENT] CALL %s %s (%x)", c.effectiveCanisterID, c.methodName, c.requestID)
	rawCertificate, err := c.a.call(c.effectiveCanisterID, c.data)
	if err != nil {
		return err
	}
	if len(rawCertificate) != 0 {
		var certificate certification.Certificate
		if err := cbor.Unmarshal(rawCertificate, &certificate); err != nil {
			return err
		}
		raw, err := hashtree.Lookup(certificate.Tree.Root, hashtree.Label("request_status"), c.requestID[:], hashtree.Label("reply"))
		if err != nil {
			return fmt.Errorf("no reply found")
		}
		return c.unmarshal(raw, out)
	}

	raw, err := c.a.poll(c.effectiveCanisterID, c.requestID)
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
