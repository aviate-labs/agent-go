package pocketic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Base64EncodedBlob is a byte slice that is base64 encoded when marshaled to JSON.
// The underlying byte slice is already decoded, so it can be used as is.
type Base64EncodedBlob []byte

func (b Base64EncodedBlob) MarshalJSON() ([]byte, error) {
	encoded := base64.StdEncoding.EncodeToString(b)
	return json.Marshal(encoded)
}

func (b Base64EncodedBlob) String() string {
	return base64.StdEncoding.EncodeToString(b)
}

func (b *Base64EncodedBlob) UnmarshalJSON(bytes []byte) error {
	var encoded string
	if err := json.Unmarshal(bytes, &encoded); err != nil {
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	*b = decoded
	return nil
}

type RawAddCycles struct {
	Amount     int               `json:"amount"`
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type RawCanisterCall struct {
	CanisterID         Base64EncodedBlob     `json:"canister_id"`
	EffectivePrincipal RawEffectivePrincipal `json:"effective_principal"`
	Method             string                `json:"method"`
	Payload            Base64EncodedBlob     `json:"payload"`
	Sender             Base64EncodedBlob     `json:"sender"`
}

func (r *RawCanisterCall) UnmarshalJSON(bytes []byte) error {
	var raw struct {
		CanisterID         Base64EncodedBlob `json:"canister_id"`
		EffectivePrincipal json.RawMessage   `json:"effective_principal"`
		Method             string            `json:"method"`
		Payload            Base64EncodedBlob `json:"payload"`
		Sender             Base64EncodedBlob `json:"sender"`
	}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}
	ep, err := unmarshalRawEffectivePrincipal(raw.EffectivePrincipal)
	if err != nil {
		return err
	}
	r.CanisterID = raw.CanisterID
	r.EffectivePrincipal = ep
	r.Method = raw.Method
	r.Payload = raw.Payload
	r.Sender = raw.Sender
	return nil
}

type RawCanisterID struct {
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type RawCanisterResult Result[RawWasmResult]

type RawEffectivePrincipal interface {
	rawEffectivePrincipal()
}

func unmarshalRawEffectivePrincipal(bytes []byte) (RawEffectivePrincipal, error) {
	var none RawEffectivePrincipalNone
	if err := json.Unmarshal(bytes, &none); err == nil {
		return none, nil
	}
	var m map[string]Base64EncodedBlob
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	if canisterID, ok := m["CanisterId"]; ok {
		return RawEffectivePrincipalCanisterID{CanisterID: canisterID}, nil
	}
	if subnetID, ok := m["SubnetId"]; ok {
		return RawEffectivePrincipalSubnetID{SubnetID: subnetID}, nil
	}
	return nil, fmt.Errorf("unknown effective principal: %s", string(bytes))
}

type RawEffectivePrincipalCanisterID struct {
	CanisterID Base64EncodedBlob `json:"CanisterId"`
}

func (RawEffectivePrincipalCanisterID) rawEffectivePrincipal() {}

type RawEffectivePrincipalNone struct{}

func (n RawEffectivePrincipalNone) MarshalJSON() ([]byte, error) {
	return json.Marshal(new(None))
}

func (n RawEffectivePrincipalNone) UnmarshalJSON(bytes []byte) error {
	var none None
	return json.Unmarshal(bytes, &none)
}

func (RawEffectivePrincipalNone) rawEffectivePrincipal() {}

type RawEffectivePrincipalSubnetID struct {
	SubnetID Base64EncodedBlob `json:"SubnetId"`
}

func (RawEffectivePrincipalSubnetID) rawEffectivePrincipal() {}

type RawMessageID struct {
	EffectivePrincipal RawEffectivePrincipal `json:"effective_principal"`
	MessageID          Base64EncodedBlob     `json:"message_id"`
}

func (r *RawMessageID) UnmarshalJSON(bytes []byte) error {
	var raw struct {
		EffectivePrincipal json.RawMessage   `json:"effective_principal"`
		MessageID          Base64EncodedBlob `json:"message_id"`
	}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}
	ep, err := unmarshalRawEffectivePrincipal(raw.EffectivePrincipal)
	if err != nil {
		return err
	}
	r.EffectivePrincipal = ep
	r.MessageID = raw.MessageID
	return nil
}

type RawSetStableMemory struct {
	BlobID     Base64EncodedBlob `json:"blob_id"`
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type RawSubmitIngressResult Result[RawMessageID]

type RawSubnetId struct {
	SubnetID Base64EncodedBlob `json:"subnet_id"`
}

type RawTime struct {
	NanosSinceEpoch int `json:"nanos_since_epoch"`
}

type RawVerifyCanisterSigArg struct {
	Message    Base64EncodedBlob `json:"msg"`
	PublicKey  Base64EncodedBlob `json:"pubkey"`
	RootPubKey Base64EncodedBlob `json:"root_pubkey"`
	Signature  Base64EncodedBlob `json:"sig"`
}

type RawWasmResult WASMResult[Base64EncodedBlob]

type Reject string

func (r Reject) Error() string {
	return string(r)
}

type Result[R any] struct {
	Ok  *R         `json:"Ok,omitempty"`
	Err *UserError `json:"Err,omitempty"`
}

type UserError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (e UserError) Error() string {
	return fmt.Sprintf("(%d) %s", e.Code, e.Description)
}

// WASMResult describes the different types that executing a WASM function in a canister can produce.
type WASMResult[Blob any] struct {
	// Raw response, returned in a successful case.
	Reply *Blob `json:"reply,omitempty"`
	// Returned with an error message when the canister decides to reject the message.
	Reject *Reject `json:"reject,omitempty"`
}
