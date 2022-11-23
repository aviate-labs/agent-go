package agent

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"math/rand"
	"sort"

	"github.com/aviate-labs/agent-go/candid"
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

func hashPaths(paths [][][]byte) [32]byte {
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

// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#http-call
type Request struct {
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
	Paths [][][]byte
}

func NewRequest(
	sender principal.Principal,
	requestType RequestType,
	canisterID principal.Principal,
	methodName string,
	arguments string,
	effective uint64,
) (Request, error) {
	args, err := candid.EncodeValue(arguments)
	if err != nil {
		return Request{}, nil
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return Request{}, err
	}
	return Request{
		Type:          requestType,
		Sender:        sender,
		Nonce:         nonce,
		IngressExpiry: 1000000000 * (effective + 300),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     args,
	}, nil
}

func (r Request) MarshalCBOR() ([]byte, error) {
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

type RequestID [32]byte

// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#request-id
func NewRequestID(req Request) RequestID {
	var hashes [][]byte
	if len(req.Type) != 0 {
		typeHash := sha256.Sum256([]byte(req.Type))
		hashes = append(hashes, append(typeKey[:], typeHash[:]...))
	}
	if len(req.CanisterID.Raw) != 0 {
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

func (r RequestID) Sign(id identity.Identity) []byte {
	return id.Sign(append(
		// \x0Aic-request
		[]byte{0x0a, 0x69, 0x63, 0x2d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74},
		r[:]...,
	))
}

type RequestType = string

const (
	RequestTypeCall      RequestType = "call"
	RequestTypeQuery     RequestType = "query"
	RequestTypeReadState RequestType = "read_state"
)

type requestRaw struct {
	Type          RequestType `cbor:"request_type"`
	Sender        []byte      `cbor:"sender"`
	Nonce         []byte      `cbor:"nonce"`
	IngressExpiry uint64      `cbor:"ingress_expiry"`
	CanisterID    []byte      `cbor:"canister_id"`
	MethodName    string      `cbor:"method_name"`
	Arguments     []byte      `cbor:"arg"`
	Paths         [][][]byte  `cbor:"paths,omitempty"`
}
