package agent

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"google.golang.org/protobuf/proto"
)

// Call calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Call(canisterID principal.Principal, methodName string, args []any, values []any) error {
	call, err := a.CreateCall(canisterID, methodName, args...)
	if err != nil {
		return err
	}
	return call.CallAndWait(values...)
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

// CreateCall creates a new Call to the given canister and method.
func (a *Agent) CreateCall(canisterID principal.Principal, methodName string, args ...any) (*Call, error) {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
		Nonce:         nonce,
	})
	if err != nil {
		return nil, err
	}
	return &Call{
		a:                   a,
		methodName:          methodName,
		effectiveCanisterID: effectiveCanisterID(canisterID, args),
		requestID:           *requestID,
		data:                data,
	}, nil
}

// Call is an intermediate representation of a Call to a canister.
type Call struct {
	a                   *Agent
	methodName          string
	effectiveCanisterID principal.Principal
	requestID           RequestID
	data                []byte
}

// Call calls a method on a canister, it does not wait for the result.
func (c Call) Call() error {
	c.a.logger.Printf("[AGENT] CALL %s %s (%x)", c.effectiveCanisterID, c.methodName, c.requestID)
	_, err := c.a.call(c.effectiveCanisterID, c.data)
	return err
}

// CallAndWait calls a method on a canister and waits for the result.
func (c Call) CallAndWait(values ...any) error {
	if err := c.Call(); err != nil {
		return err
	}
	return c.Wait(values...)
}

// Wait waits for the result of the Call and unmarshals it into the given values.
func (c Call) Wait(values ...any) error {
	raw, err := c.a.poll(c.effectiveCanisterID, c.requestID)
	if err != nil {
		return err
	}
	return idl.Unmarshal(raw, values)
}

// WithEffectiveCanisterID sets the effective canister ID for the Call.
func (c *Call) WithEffectiveCanisterID(canisterID principal.Principal) *Call {
	c.effectiveCanisterID = canisterID
	return c
}
