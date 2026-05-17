package identity

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestNewSecp256k1Identity(t *testing.T) {
	id, _ := NewRandomSecp256k1Identity()
	data, err := id.ToPEM()
	if err != nil {
		t.Fatal(err)
	}
	id_, err := NewSecp256k1IdentityFromPEM(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.privateKey.Bytes(), id_.privateKey.Bytes()) {
		t.Error()
	}
	if !bytes.Equal(id.PublicKey(), id_.PublicKey()) {
		t.Error()
	}
}

// TestSecp256k1Identity_PEMRoundTrip_KnownVector loads a known PEM, signs a
// known message, then verifies. Also confirms that ToPEM serialises back to
// something that parses to the same public key. Acts as a tripwire if the
// secp256k1 implementation backing this package changes.
//
// The PEM below is a standard SEC1 EC private key on secp256k1. Reproduce via:
//
//	openssl ec -in <pem> -text -noout
//
// Decoded contents:
//
//	curve  = secp256k1 (OID 1.3.132.0.10, DER: 06 05 2b 81 04 00 0a)
//	scalar = 0832ee76 4471 51e4 4386 7529 da9b cbc4
//	         b0c8 08b2 2834 26b5 b107 4c83 3e58 d781
//	pub.X  = 80ef3bac 9d68 cf37 4cbc 9c99 43e1 8004
//	         3a94 c462 ef82 7027 4e57 089d 5dcd ba1b
//	pub.Y  = 8fbf83b7 546c cc1b 3781 3777 76e9 c471
//	         0fdf 533e d9bc ba8d 8ebb 32a1 4a2a ba11
//
// wantDER is the SubjectPublicKeyInfo (RFC 5480) wrapper that
// derEncodeSecp256k1PublicKey emits: ASN.1 SEQUENCE of {algorithm = (ecPublicKey,
// secp256k1), publicKey = uncompressed point 0x04 || X || Y}.
func TestSecp256k1Identity_PEMRoundTrip_KnownVector(t *testing.T) {
	pem := `
-----BEGIN EC PARAMETERS-----
BgUrgQQACg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHQCAQEEIAgy7nZEcVHkQ4Z1Kdqby8SwyAiyKDQmtbEHTIM+WNeBoAcGBSuBBAAK
oUQDQgAEgO87rJ1ozzdMvJyZQ+GABDqUxGLvgnAnTlcInV3NuhuPv4O3VGzMGzeB
N3d26cRxD99TPtm8uo2OuzKhSiq6EQ==
-----END EC PRIVATE KEY-----
`
	wantDER, _ := hex.DecodeString("3056301006072a8648ce3d020106052b8104000a0342000480ef3bac9d68cf374cbc9c9943e180043a94c462ef8270274e57089d5dcdba1b8fbf83b7546ccc1b3781377776e9c4710fdf533ed9bcba8d8ebb32a14a2aba11")
	id, err := NewSecp256k1IdentityFromPEM([]byte(pem))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.PublicKey(), wantDER) {
		t.Fatal("public key mismatch on load")
	}
	msg := []byte("ic agent secp256k1 signature test vector")
	sig, err := id.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) != 64 {
		t.Fatalf("signature length: got %d, want 64", len(sig))
	}
	if !id.Verify(msg, sig) {
		t.Fatal("self-verify failed")
	}
	out, err := id.ToPEM()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := NewSecp256k1IdentityFromPEM(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.PublicKey(), id2.PublicKey()) {
		t.Fatal("public key mismatch on ToPEM round-trip")
	}
	if !id2.Verify(msg, sig) {
		t.Fatal("re-loaded identity could not verify original signature")
	}
}

// TestSecp256k1Identity_PEM_NoEmbeddedPublicKey covers SEC1 PEMs whose
// ECPrivateKey omits the OPTIONAL publicKey field (RFC 5915). Produced via
// `openssl ec -in <pem> -no_public`. Both forms must yield the same identity.
func TestSecp256k1Identity_PEM_NoEmbeddedPublicKey(t *testing.T) {
	withPub := `-----BEGIN EC PRIVATE KEY-----
MHQCAQEEICR11qAER72cCYEm80d7AC6IbO7l5nH8epO4wk0IXZc2oAcGBSuBBAAK
oUQDQgAEd6GaY6fx7WJAE/x/F1+m+nI7EgdQSYTojnfxOn0nAXwuFbwufil3N0rD
OTpP5OZ73DxVTd4zsf4So69iMI0sOQ==
-----END EC PRIVATE KEY-----
`
	noPub := `-----BEGIN EC PRIVATE KEY-----
MC4CAQEEICR11qAER72cCYEm80d7AC6IbO7l5nH8epO4wk0IXZc2oAcGBSuBBAAK
-----END EC PRIVATE KEY-----
`
	idA, err := NewSecp256k1IdentityFromPEMWithoutParameters([]byte(withPub))
	if err != nil {
		t.Fatal(err)
	}
	idB, err := NewSecp256k1IdentityFromPEMWithoutParameters([]byte(noPub))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(idA.PublicKey(), idB.PublicKey()) {
		t.Fatal("public key mismatch between embedded and derived")
	}
	msg := []byte("derived-pubkey signing path")
	sig, err := idB.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !idB.Verify(msg, sig) {
		t.Fatal("derived-pubkey identity failed self-verify")
	}
}

func TestSecp256k1Identity_Sign(t *testing.T) {
	id, err := NewRandomSecp256k1Identity()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello")
	sig, err := id.Sign(data)
	if err != nil {
		t.Fatal(err)
	}
	if !id.Verify(data, sig) {
		t.Error()
	}
}

// TestSecp256k1Identity_SignatureFormat pins the on-wire signature shape:
// 64 bytes total, r and s each 32 bytes big-endian zero-padded.
func TestSecp256k1Identity_SignatureFormat(t *testing.T) {
	id, err := NewRandomSecp256k1Identity()
	if err != nil {
		t.Fatal(err)
	}
	sig, err := id.Sign([]byte("anything"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) != 64 {
		t.Fatalf("signature length: got %d, want 64", len(sig))
	}
}
