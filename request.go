package agent

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"sort"

	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"

	"github.com/fxamacker/cbor/v2"
)

var (
	typeKey          = sha256.Sum256([]byte("request_type"))
	canisterIDKey    = sha256.Sum256([]byte("canister_id"))
	nonceKey         = sha256.Sum256([]byte("nonce"))
	methodNameKey    = sha256.Sum256([]byte("method_name"))
	argumentsKey     = sha256.Sum256([]byte("arg"))
	ingressExpiryKey = sha256.Sum256([]byte("ingress_expiry"))
	senderKey        = sha256.Sum256([]byte("sender"))
	pathsKey         = sha256.Sum256([]byte("paths"))
)

func encodeLEB128(i uint64) []byte {
	bi := big.NewInt(int64(i))
	e, _ := leb128.EncodeUnsigned(bi)
	return e
}

func hashPaths(paths [][]hashtree.Label) [32]byte {
	var hash []byte
	for _, path := range paths {
		var rawPathHash []byte
		for _, p := range path {
			pathBytes := sha256.Sum256(p)
			rawPathHash = append(rawPathHash, pathBytes[:]...)
		}
		pathHash := sha256.Sum256(rawPathHash)
		hash = append(hash, pathHash[:]...)
	}
	return sha256.Sum256(hash)
}

// Request is the request to the agent.
// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#http-call
type Request struct {
	// The type of the request. This is used to distinguish between query, call and read_state requests.
	Type RequestType
	// The user who issued the request.
	Sender principal.Principal
	// Arbitrary user-provided data, typically randomly generated. This can be
	// used to create distinct requests with otherwise identical fields.
	Nonce []byte
	// An upper limit on the validity of the request, expressed in nanoseconds
	// since 1970-01-01 (like ic0.time()).
	IngressExpiry uint64
	// The principal of the canister to call.
	CanisterID principal.Principal
	// Name of the canister method to call.
	MethodName string
	// Argument to pass to the canister method.
	Arguments []byte
	// A list of paths, where a path is itself a sequence of blobs.
	Paths [][]hashtree.Label
}

// MarshalCBOR implements the CBOR marshaler interface.
func (r *Request) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(requestRaw{
		Type:          r.Type,
		Sender:        r.Sender.Raw,
		Nonce:         r.Nonce,
		IngressExpiry: r.IngressExpiry,
		CanisterID:    r.CanisterID.Raw,
		MethodName:    r.MethodName,
		Arguments:     r.Arguments,
		Paths:         r.Paths,
	})
}

// UnmarshalCBOR implements the CBOR unmarshaler interface.
func (r *Request) UnmarshalCBOR(data []byte) error {
	var raw requestRaw
	if err := cbor.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Type = raw.Type
	r.Sender = principal.Principal{Raw: raw.Sender}
	r.Nonce = raw.Nonce
	r.IngressExpiry = raw.IngressExpiry
	r.CanisterID = principal.Principal{Raw: raw.CanisterID}
	r.MethodName = raw.MethodName
	r.Arguments = raw.Arguments
	r.Paths = raw.Paths
	return nil
}

// RequestID is the request ID.
type RequestID [32]byte

// NewRequestID creates a new request ID.
// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#request-id
func NewRequestID(req Request) RequestID {
	var hashes [][]byte
	if len(req.Type) != 0 {
		typeHash := sha256.Sum256([]byte(req.Type))
		hashes = append(hashes, append(typeKey[:], typeHash[:]...))
	}
	// NOTE: the canister ID may be the empty slice. The empty slice doesn't mean it's not
	// set, it means it's the management canister (aaaaa-aa).
	if req.CanisterID.Raw != nil {
		canisterIDHash := sha256.Sum256(req.CanisterID.Raw)
		hashes = append(hashes, append(canisterIDKey[:], canisterIDHash[:]...))
	}
	if len(req.MethodName) != 0 {
		methodNameHash := sha256.Sum256([]byte(req.MethodName))
		hashes = append(hashes, append(methodNameKey[:], methodNameHash[:]...))
	}
	if len(req.Arguments) != 0 {
		argumentsHash := sha256.Sum256(req.Arguments)
		hashes = append(hashes, append(argumentsKey[:], argumentsHash[:]...))
	}
	if len(req.Sender.Raw) != 0 {
		senderHash := sha256.Sum256(req.Sender.Raw)
		hashes = append(hashes, append(senderKey[:], senderHash[:]...))
	}
	if req.IngressExpiry != 0 {
		ingressExpiryHash := sha256.Sum256(encodeLEB128(req.IngressExpiry))
		hashes = append(hashes, append(ingressExpiryKey[:], ingressExpiryHash[:]...))
	}
	if len(req.Nonce) != 0 {
		nonceHash := sha256.Sum256(req.Nonce)
		hashes = append(hashes, append(nonceKey[:], nonceHash[:]...))
	}
	if req.Paths != nil {
		pathsHash := hashPaths(req.Paths)
		hashes = append(hashes, append(pathsKey[:], pathsHash[:]...))
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) == -1
	})
	return sha256.Sum256(bytes.Join(hashes, nil))
}

// Sign signs the request ID with the given identity.
func (r RequestID) Sign(id identity.Identity) []byte {
	message := append(
		// \x0Aic-request
		[]byte{0x0a, 0x69, 0x63, 0x2d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74},
		r[:]...,
	)
	signature := id.Sign(message)
	return signature
}

// RequestType is the type of request.
type RequestType = string

const (
	// RequestTypeCall is a call request.
	RequestTypeCall RequestType = "call"
	// RequestTypeQuery is a query request.
	RequestTypeQuery RequestType = "query"
	// RequestTypeReadState is a read state request.
	RequestTypeReadState RequestType = "read_state"
)

type requestRaw struct {
	Type          RequestType        `cbor:"request_type"`
	Sender        []byte             `cbor:"sender"`
	Nonce         []byte             `cbor:"nonce"`
	IngressExpiry uint64             `cbor:"ingress_expiry"`
	CanisterID    []byte             `cbor:"canister_id"`
	MethodName    string             `cbor:"method_name"`
	Arguments     []byte             `cbor:"arg"`
	Paths         [][]hashtree.Label `cbor:"paths,omitempty"`
}
