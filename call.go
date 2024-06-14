package agent

import (
	"github.com/aviate-labs/agent-go/principal"
	"google.golang.org/protobuf/proto"
)

// Call calls a method on a canister, it does not wait for the result.
func (c APIRequest[_, _]) Call() error {
	c.a.logger.Printf("[AGENT] CALL %s %s (%x)", c.effectiveCanisterID, c.methodName, c.requestID)
	_, err := c.a.call(c.effectiveCanisterID, c.data)
	return err
}

// CallAndWait calls a method on a canister and waits for the result.
func (c APIRequest[_, Out]) CallAndWait(out Out) error {
	if err := c.Call(); err != nil {
		return err
	}
	return c.Wait(out)
}

// Wait waits for the result of the Call and unmarshals it into the given values.
func (c APIRequest[_, Out]) Wait(out Out) error {
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
	payload, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		IngressExpiry: a.expiryDate(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     payload,
	})
	if err != nil {
		return err
	}
	if _, err := a.call(canisterID, data); err != nil {
		return err
	}
	raw, err := a.poll(canisterID, *requestID)
	if err != nil {
		return err
	}
	return proto.Unmarshal(raw, out)
}
