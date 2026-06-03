package agent

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"sort"
	"testing"

	"github.com/niccolofant/agent-go/leb128"
	"github.com/niccolofant/agent-go/principal"
)

// referenceRequestID is the previous append/bytes.Join/sort.Slice/big+leb128
// implementation, kept here verbatim so the optimized NewRequestID can be
// proven byte-identical across many inputs (not just the spec known answers).
func referenceRequestID(req Request) RequestID {
	var hashes [][]byte
	if len(req.Type) != 0 {
		h := sha256.Sum256([]byte(req.Type))
		hashes = append(hashes, append(typeKey[:], h[:]...))
	}
	if req.CanisterID.Raw != nil {
		h := sha256.Sum256(req.CanisterID.Raw)
		hashes = append(hashes, append(canisterIDKey[:], h[:]...))
	}
	if len(req.MethodName) != 0 {
		h := sha256.Sum256([]byte(req.MethodName))
		hashes = append(hashes, append(methodNameKey[:], h[:]...))
	}
	if req.Arguments != nil {
		h := sha256.Sum256(req.Arguments)
		hashes = append(hashes, append(argumentsKey[:], h[:]...))
	}
	if len(req.Sender.Raw) != 0 {
		h := sha256.Sum256(req.Sender.Raw)
		hashes = append(hashes, append(senderKey[:], h[:]...))
	}
	if req.IngressExpiry != 0 {
		bi := big.NewInt(int64(req.IngressExpiry))
		e, _ := leb128.EncodeUnsigned(bi)
		h := sha256.Sum256(e)
		hashes = append(hashes, append(ingressExpiryKey[:], h[:]...))
	}
	if len(req.Nonce) != 0 {
		h := sha256.Sum256(req.Nonce)
		hashes = append(hashes, append(nonceKey[:], h[:]...))
	}
	if req.Paths != nil {
		h := hashPaths(req.Paths)
		hashes = append(hashes, append(pathsKey[:], h[:]...))
	}
	sort.Slice(hashes, func(i, j int) bool { return bytes.Compare(hashes[i], hashes[j]) == -1 })
	return sha256.Sum256(bytes.Join(hashes, nil))
}

func TestNewRequestIDMatchesReference(t *testing.T) {
	cases := []Request{
		// Typical update call.
		{
			Type:          RequestTypeCall,
			Sender:        principal.Principal{Raw: []byte{1, 2, 3, 4}},
			CanisterID:    principal.Principal{Raw: []byte{0xab, 0xcd}},
			MethodName:    "swap",
			Arguments:     []byte{0x44, 0x49, 0x44, 0x4c},
			IngressExpiry: 1_700_000_000_000_000_000,
		},
		// Management canister: EMPTY but non-nil CanisterID must stay in the id.
		{
			Type:          RequestTypeCall,
			Sender:        principal.Principal{Raw: []byte{9}},
			CanisterID:    principal.Principal{Raw: []byte{}},
			MethodName:    "provisional_create_canister_with_cycles",
			Arguments:     []byte{},
			IngressExpiry: 1,
		},
		// Read state (paths, no method/canister).
		{
			Type:          RequestTypeReadState,
			Sender:        principal.Principal{Raw: []byte{7, 7}},
			IngressExpiry: 42,
		},
		// Query with a nonce.
		{
			Type:          RequestTypeQuery,
			Sender:        principal.Principal{Raw: []byte{5}},
			CanisterID:    principal.Principal{Raw: []byte{6, 7, 8}},
			MethodName:    "metadata",
			Arguments:     []byte{0},
			Nonce:         []byte{0xde, 0xad, 0xbe, 0xef},
			IngressExpiry: 999_999_999,
		},
		// Edge ingress-expiry values that exercise ULEB128 boundaries.
		{Type: RequestTypeCall, Sender: principal.Principal{Raw: []byte{1}}, IngressExpiry: 127},
		{Type: RequestTypeCall, Sender: principal.Principal{Raw: []byte{1}}, IngressExpiry: 128},
		{Type: RequestTypeCall, Sender: principal.Principal{Raw: []byte{1}}, IngressExpiry: 16383},
		{Type: RequestTypeCall, Sender: principal.Principal{Raw: []byte{1}}, IngressExpiry: 16384},
		// Minimal (only one field).
		{Type: RequestTypeCall},
	}
	for i, req := range cases {
		got := NewRequestID(req)
		want := referenceRequestID(req)
		if got != want {
			t.Fatalf("case %d: NewRequestID = %x, reference = %x", i, got, want)
		}
	}
}

func TestNewRequestIDZeroAlloc(t *testing.T) {
	req := Request{
		Type:          RequestTypeCall,
		Sender:        principal.Principal{Raw: []byte{1, 2, 3, 4}},
		CanisterID:    principal.Principal{Raw: []byte{0xab, 0xcd}},
		MethodName:    "swap",
		Arguments:     []byte{0x44, 0x49, 0x44, 0x4c},
		IngressExpiry: 1_700_000_000_000_000_000,
	}
	allocs := testing.AllocsPerRun(100, func() { _ = NewRequestID(req) })
	if allocs != 0 {
		t.Fatalf("NewRequestID allocations = %v, want 0", allocs)
	}
}
