package http

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/bls"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/certification/http/certexp"
	"github.com/aviate-labs/agent-go/leb128"
	"github.com/aviate-labs/agent-go/principal"
	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/fxamacker/cbor/v2"
)

const noHeadersCertExpr = `default_certification(ValidationArgs{certification:Certification{no_request_certification: Empty{},response_certification:ResponseCertification{response_header_exclusions:ResponseHeaderList{headers:[]}}}})`

var fakeCanister = principal.Principal{Raw: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01}}

func TestVerifier_V1_FallbackRejectsMismatchedBody(t *testing.T) {
	indexBody := []byte("<html></html>")
	indexHash := sha256.Sum256(indexBody)
	canisterTree := hashtree.Labeled{
		Label: hashtree.Label("http_assets"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("/index.html"),
			Tree:  hashtree.Leaf(indexHash[:]),
		},
	}
	cert, derRootKey := buildSignedCertificate(t, fakeCanister, canisterTree.Reconstruct())
	header := buildICCertificateHeader(t, cert, canisterTree, 1, nil)

	resp := &Response{
		StatusCode: 200,
		Headers:    []HeaderField{{Field0: "IC-Certificate", Field1: header}},
		Body:       []byte("an entirely different body that was not witnessed"),
	}
	v := Verifier{CanisterID: fakeCanister, RootKey: derRootKey}
	if err := v.Verify("/missing.txt", nil, resp); err == nil {
		t.Fatal("expected fallback to reject body that does not match witnessed /index.html hash")
	}
}

func TestVerifier_V1_FallbackToIndexHTML(t *testing.T) {
	body := []byte("<html></html>")
	bodyHash := sha256.Sum256(body)
	canisterTree := hashtree.Labeled{
		Label: hashtree.Label("http_assets"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("/index.html"),
			Tree:  hashtree.Leaf(bodyHash[:]),
		},
	}
	cert, derRootKey := buildSignedCertificate(t, fakeCanister, canisterTree.Reconstruct())
	header := buildICCertificateHeader(t, cert, canisterTree, 1, nil)

	resp := &Response{
		StatusCode: 200,
		Headers:    []HeaderField{{Field0: "IC-Certificate", Field1: header}},
		Body:       body,
	}
	v := Verifier{CanisterID: fakeCanister, RootKey: derRootKey}
	if err := v.Verify("/missing.txt", nil, resp); err != nil {
		t.Fatalf("expected fallback to /index.html witness to succeed, got: %v", err)
	}
}

func TestVerifier_V1_HappyPath(t *testing.T) {
	body := []byte("hello world")
	bodyHash := sha256.Sum256(body)
	canisterTree := hashtree.Labeled{
		Label: hashtree.Label("http_assets"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("/index.html"),
			Tree:  hashtree.Leaf(bodyHash[:]),
		},
	}
	cert, derRootKey := buildSignedCertificate(t, fakeCanister, canisterTree.Reconstruct())
	header := buildICCertificateHeader(t, cert, canisterTree, 1, nil)

	resp := &Response{
		StatusCode: 200,
		Headers:    []HeaderField{{Field0: "IC-Certificate", Field1: header}},
		Body:       body,
	}
	v := Verifier{CanisterID: fakeCanister, RootKey: derRootKey}
	if err := v.Verify("/index.html", nil, resp); err != nil {
		t.Fatalf("verify legit V1 response: %v", err)
	}
}

func TestVerifier_V1_TamperedTree_Rejected(t *testing.T) {
	body := []byte("hello world")
	bodyHash := sha256.Sum256(body)
	legitTree := hashtree.Labeled{
		Label: hashtree.Label("http_assets"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("/index.html"),
			Tree:  hashtree.Leaf(bodyHash[:]),
		},
	}
	cert, derRootKey := buildSignedCertificate(t, fakeCanister, legitTree.Reconstruct())
	tamperedTree := hashtree.Fork{
		LeftTree: legitTree,
		RightTree: hashtree.Labeled{
			Label: hashtree.Label("extra"),
			Tree:  hashtree.Leaf([]byte("evidence of tampering")),
		},
	}
	header := buildICCertificateHeader(t, cert, tamperedTree, 1, nil)

	resp := &Response{
		StatusCode: 200,
		Headers:    []HeaderField{{Field0: "IC-Certificate", Field1: header}},
		Body:       body,
	}
	v := Verifier{CanisterID: fakeCanister, RootKey: derRootKey}
	if err := v.Verify("/index.html", nil, resp); err == nil {
		t.Fatal("expected verification to reject tampered V1 tree, but it passed")
	}
}

// TestVerifier_V2_TamperedTree_Rejected covers the V2 binding check: the
// canister-side tree's root must match the certified_data leaf. Every other
// invariant (cert-expr hash, witnessed (req, resp) pair, recomputed
// respHash) is satisfied so the test isolates this one check.
func TestVerifier_V2_TamperedTree_Rejected(t *testing.T) {
	body := []byte("v2 body")
	resp := &Response{
		StatusCode: 200,
		Headers: []HeaderField{
			{Field0: "IC-Certificate", Field1: ""},
			{Field0: "IC-CertificateExpression", Field1: noHeadersCertExpr},
		},
		Body: body,
	}

	parsedExpr, err := certexp.ParseCertificateExpression(noHeadersCertExpr)
	if err != nil {
		t.Fatalf("parse cert expression: %v", err)
	}
	respHash, err := CalculateResponseHash(resp, parsedExpr.Certification.ResponseCertification)
	if err != nil {
		t.Fatalf("calculate response hash: %v", err)
	}

	certExprHash := sha256.Sum256([]byte(noHeadersCertExpr))
	canisterTree := hashtree.Labeled{
		Label: hashtree.Label("http_expr"),
		Tree: hashtree.Labeled{
			Label: hashtree.Label("<*>"),
			Tree: hashtree.Labeled{
				Label: certExprHash[:],
				Tree: hashtree.Labeled{
					Label: hashtree.Label(""),
					Tree: hashtree.Labeled{
						Label: respHash[:],
						Tree:  hashtree.Leaf(nil),
					},
				},
			},
		},
	}

	someOtherRoot := sha256.Sum256([]byte("not the tree the canister served"))
	cert, derRootKey := buildSignedCertificate(t, fakeCanister, someOtherRoot)
	resp.Headers[0].Field1 = buildICCertificateHeader(t, cert, canisterTree, 2, []string{"http_expr", "<*>"})

	v := Verifier{CanisterID: fakeCanister, RootKey: derRootKey}
	if err := v.Verify("/", nil, resp); err == nil {
		t.Fatal("expected rejection: canister tree root does not match certified_data")
	}
}

func buildICCertificateHeader(t *testing.T, cert certification.Certificate, canisterTree hashtree.Node, version int, exprPath []string) string {
	t.Helper()
	certCBOR, err := cbor.Marshal(cert)
	if err != nil {
		t.Fatalf("marshal cert: %v", err)
	}
	treeCBOR, err := hashtree.Serialize(canisterTree)
	if err != nil {
		t.Fatalf("marshal tree: %v", err)
	}
	parts := []string{
		fmt.Sprintf("certificate=:%s:", base64.StdEncoding.EncodeToString(certCBOR)),
		fmt.Sprintf("tree=:%s:", base64.StdEncoding.EncodeToString(treeCBOR)),
		fmt.Sprintf("version=%d", version),
	}
	if exprPath != nil {
		pathCBOR, err := cbor.Marshal(exprPath)
		if err != nil {
			t.Fatalf("marshal expr_path: %v", err)
		}
		parts = append(parts, fmt.Sprintf("expr_path=:%s:", base64.StdEncoding.EncodeToString(pathCBOR)))
	}
	return strings.Join(parts, ", ")
}

func buildSignedCertificate(t *testing.T, canisterID principal.Principal, certifiedData [32]byte) (certification.Certificate, []byte) {
	t.Helper()
	sk := bls.NewSecretKeyByCSPRNG()
	if sk == nil {
		t.Fatal("bls: failed to generate secret key")
	}
	pubAffine := bls12381.G2Affine(*sk.PublicKey())
	pubBytes := pubAffine.Bytes()
	derPub, err := certification.PublicBLSKeyToDER(pubBytes[:])
	if err != nil {
		t.Fatalf("der-encode bls pubkey: %v", err)
	}

	nowEnc, err := leb128.EncodeUnsigned(big.NewInt(time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("leb128 encode time: %v", err)
	}

	tree := hashtree.Fork{
		LeftTree: hashtree.Labeled{
			Label: hashtree.Label("canister"),
			Tree: hashtree.Labeled{
				Label: canisterID.Raw,
				Tree: hashtree.Labeled{
					Label: hashtree.Label("certified_data"),
					Tree:  hashtree.Leaf(certifiedData[:]),
				},
			},
		},
		RightTree: hashtree.Labeled{
			Label: hashtree.Label("time"),
			Tree:  hashtree.Leaf(nowEnc),
		},
	}

	root := tree.Reconstruct()
	msg := append(hashtree.DomainSeparator("ic-state-root"), root[:]...)
	sig, err := sk.Sign(msg)
	if err != nil {
		t.Fatalf("bls sign: %v", err)
	}
	sigAffine := bls12381.G1Affine(*sig)
	sigBytes := sigAffine.Bytes()

	return certification.Certificate{
		Tree:      hashtree.NewHashTree(tree),
		Signature: sigBytes[:],
	}, derPub
}
