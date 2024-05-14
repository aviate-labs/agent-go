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

// String returns the base64 encoded string of the blob.
// NOTE: it will truncate the string if it is too long.
func (b Base64EncodedBlob) String() string {
	str := base64.StdEncoding.EncodeToString(b)
	if len(str) > 20 {
		return str[:10] + "..." + str[len(str)-10:]
	}
	return str
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

type EffectivePrincipal interface {
	rawEffectivePrincipal()
}

func unmarshalRawEffectivePrincipal(bytes []byte) (EffectivePrincipal, error) {
	var none EffectivePrincipalNone
	if err := json.Unmarshal(bytes, &none); err == nil {
		return none, nil
	}
	var m map[string]Base64EncodedBlob
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	if canisterID, ok := m["CanisterId"]; ok {
		return EffectivePrincipalCanisterID{CanisterID: canisterID}, nil
	}
	if subnetID, ok := m["SubnetId"]; ok {
		return EffectivePrincipalSubnetID{SubnetID: subnetID}, nil
	}
	return nil, fmt.Errorf("unknown effective principal: %s", string(bytes))
}

type EffectivePrincipalCanisterID struct {
	CanisterID Base64EncodedBlob `json:"CanisterId"`
}

func (EffectivePrincipalCanisterID) rawEffectivePrincipal() {}

type EffectivePrincipalNone struct{}

func (n EffectivePrincipalNone) MarshalJSON() ([]byte, error) {
	return json.Marshal(new(None))
}

func (n EffectivePrincipalNone) UnmarshalJSON(bytes []byte) error {
	var none None
	return json.Unmarshal(bytes, &none)
}

func (EffectivePrincipalNone) rawEffectivePrincipal() {}

type EffectivePrincipalSubnetID struct {
	SubnetID Base64EncodedBlob `json:"SubnetId"`
}

func (EffectivePrincipalSubnetID) rawEffectivePrincipal() {}

type MessageID struct {
	EffectivePrincipal EffectivePrincipal `json:"effective_principal"`
	MessageID          Base64EncodedBlob  `json:"message_id"`
}

func (r *MessageID) UnmarshalJSON(bytes []byte) error {
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

type RawCanisterID struct {
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type Reject string

func (r Reject) Error() string {
	return string(r)
}

type SubnetID struct {
	SubnetID Base64EncodedBlob `json:"subnet_id"`
}

type UserError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e UserError) Error() string {
	return fmt.Sprintf("(%s) %s", e.Code, e.Description)
}

type VerifyCanisterSigArg struct {
	Message    Base64EncodedBlob `json:"msg"`
	PublicKey  Base64EncodedBlob `json:"pubkey"`
	RootPubKey Base64EncodedBlob `json:"root_pubkey"`
	Signature  Base64EncodedBlob `json:"sig"`
}

type createResponse[T any] struct {
	Created *T            `json:"Created"`
	Error   *ErrorMessage `json:"Error"`
}

type rawAddCycles struct {
	Amount     int               `json:"amount"`
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type rawCanisterCall struct {
	CanisterID         Base64EncodedBlob  `json:"canister_id"`
	EffectivePrincipal EffectivePrincipal `json:"effective_principal"`
	Method             string             `json:"method"`
	Payload            Base64EncodedBlob  `json:"payload"`
	Sender             Base64EncodedBlob  `json:"sender"`
}

func (r *rawCanisterCall) UnmarshalJSON(bytes []byte) error {
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

type rawCanisterResult result[rawWasmResult]

type rawCycles struct {
	Cycles int `json:"cycles"`
}

type rawSetStableMemory struct {
	BlobID     Base64EncodedBlob `json:"blob_id"`
	CanisterID Base64EncodedBlob `json:"canister_id"`
}

type rawStableMemory struct {
	Blob Base64EncodedBlob `json:"blob"`
}

type rawSubmitIngressResult result[MessageID]

type rawTime struct {
	NanosSinceEpoch int64 `json:"nanos_since_epoch"`
}

type rawWasmResult wasmResult[Base64EncodedBlob]

type result[R any] struct {
	Ok  *R         `json:"Ok,omitempty"`
	Err *UserError `json:"Err,omitempty"`
}

// wasmResult describes the different types that executing a WASM function in a canister can produce.
type wasmResult[Blob any] struct {
	// Raw response, returned in a successful case.
	Reply *Blob `json:"reply,omitempty"`
	// Returned with an error message when the canister decides to reject the message.
	Reject *Reject `json:"reject,omitempty"`
}
