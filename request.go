package agent

import (
	"bytes"
	"crypto/sha256"

	"github.com/niccolofant/agent-go/certification/hashtree"
	"github.com/niccolofant/agent-go/identity"
	"github.com/niccolofant/agent-go/principal"

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

func hashPaths(paths [][]hashtree.Label) [32]byte {
	hash := make([]byte, 0, len(paths)*sha256.Size)
	for _, path := range paths {
		rawPathHash := make([]byte, 0, len(path)*sha256.Size)
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
	m := make(map[string]any)
	if len(r.Type) != 0 {
		m["request_type"] = r.Type
	}
	if r.CanisterID.Raw != nil {
		m["canister_id"] = r.CanisterID.Raw
	}
	if len(r.MethodName) != 0 {
		m["method_name"] = r.MethodName
	}
	if r.Arguments != nil {
		// Some endpoints require the argument to be an empty array, not null.
		// This is the case with the protobuf endpoints on the registry.
		m["arg"] = r.Arguments
	}
	if len(r.Sender.Raw) != 0 {
		m["sender"] = r.Sender.Raw
	}
	if r.IngressExpiry != 0 {
		m["ingress_expiry"] = r.IngressExpiry
	}
	if len(r.Nonce) != 0 {
		m["nonce"] = r.Nonce
	}
	if r.Paths != nil {
		m["paths"] = r.Paths
	}
	return cbor.Marshal(m)
}

// RequestID is the request ID.
type RequestID [32]byte

// requestFields is the maximum number of fields a request id is built from
// (type, canister_id, method_name, arg, sender, ingress_expiry, nonce, paths).
const requestFields = 8

// rowBytes is the size of one hashed field entry: sha256(key) || sha256(value).
const rowBytes = 64

// NewRequestID creates a new request ID.
// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#request-id
//
// Hot path: this runs on every signed request (queries, reads, and update
// calls). It is written to allocate nothing — the per-field entries live in a
// fixed stack buffer, are sorted in place (≤8 elements), and hashed directly —
// versus the previous append/bytes.Join/sort.Slice version that did ~10 heap
// allocations per call. The computed id is byte-for-byte identical (verified by
// the spec known-answer tests in request_test.go).
func NewRequestID(req Request) RequestID {
	var flat [requestFields * rowBytes]byte
	n := 0

	if len(req.Type) != 0 {
		putRow(flat[:], n, typeKey, sha256.Sum256([]byte(req.Type)))
		n++
	}
	// NOTE: the canister ID may be the empty slice. The empty slice doesn't mean
	// it's not set, it means it's the management canister (aaaaa-aa) — so this
	// keys off Raw != nil, NOT len, to keep that case in the id.
	if req.CanisterID.Raw != nil {
		putRow(flat[:], n, canisterIDKey, sha256.Sum256(req.CanisterID.Raw))
		n++
	}
	if len(req.MethodName) != 0 {
		putRow(flat[:], n, methodNameKey, sha256.Sum256([]byte(req.MethodName)))
		n++
	}
	if req.Arguments != nil {
		putRow(flat[:], n, argumentsKey, sha256.Sum256(req.Arguments))
		n++
	}
	if len(req.Sender.Raw) != 0 {
		putRow(flat[:], n, senderKey, sha256.Sum256(req.Sender.Raw))
		n++
	}
	if req.IngressExpiry != 0 {
		var lebBuf [10]byte // a uint64 ULEB128 is at most 10 bytes
		putRow(flat[:], n, ingressExpiryKey, sha256.Sum256(uleb128(req.IngressExpiry, lebBuf[:])))
		n++
	}
	if len(req.Nonce) != 0 {
		putRow(flat[:], n, nonceKey, sha256.Sum256(req.Nonce))
		n++
	}
	if req.Paths != nil {
		putRow(flat[:], n, pathsKey, hashPaths(req.Paths))
		n++
	}

	// Representation-independent request-id hashing: sort the (key||valueHash)
	// rows lexicographically, then hash their concatenation. Insertion sort —
	// n ≤ 8 and it's alloc-free (sort.Slice escapes a closure + reflect.Swapper).
	sortRows(flat[:], n)
	return sha256.Sum256(flat[:n*rowBytes])
}

// putRow writes sha256(key) || valueHash into row n of flat.
func putRow(flat []byte, n int, key, valueHash [32]byte) {
	off := n * rowBytes
	copy(flat[off:off+32], key[:])
	copy(flat[off+32:off+64], valueHash[:])
}

// sortRows sorts the first n 64-byte rows of flat lexicographically, in place.
func sortRows(flat []byte, n int) {
	for i := 1; i < n; i++ {
		for j := i; j > 0; j-- {
			a := flat[(j-1)*rowBytes : j*rowBytes]
			b := flat[j*rowBytes : (j+1)*rowBytes]
			if bytes.Compare(b, a) >= 0 {
				break
			}
			var tmp [rowBytes]byte
			copy(tmp[:], a)
			copy(a, b)
			copy(b, tmp[:])
		}
	}
}

// uleb128 encodes v as unsigned LEB128 into buf, returning the used prefix.
// Equivalent to leb128.EncodeUnsigned(big.NewInt(int64(v))) for any uint64, but
// allocation-free. (kept identical to the spec's canonical ULEB128.)
func uleb128(v uint64, buf []byte) []byte {
	i := 0
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		buf[i] = b
		i++
		if v == 0 {
			return buf[:i]
		}
	}
}

// Sign signs the request ID with the given identity.
func (r RequestID) Sign(id identity.Identity) ([]byte, error) {
	message := append(
		// \x0Aic-request
		[]byte{0x0a, 0x69, 0x63, 0x2d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74},
		r[:]...,
	)
	return id.Sign(message)
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
