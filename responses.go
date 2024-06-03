package agent

import (
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

// Response is the response from the agent.
type Response struct {
	Status     string              `cbor:"status"`
	Reply      cbor.RawMessage     `cbor:"reply"`
	RejectCode uint64              `cbor:"reject_code"`
	RejectMsg  string              `cbor:"reject_message"`
	ErrorCode  string              `cbor:"error_code"`
	Signatures []ResponseSignature `cbor:"signatures"`
}

type ResponseSignature struct {
	Timestamp int64               `cbor:"timestamp"`
	Signature []byte              `cbor:"signature"`
	Identity  principal.Principal `cbor:"identity"`
}
