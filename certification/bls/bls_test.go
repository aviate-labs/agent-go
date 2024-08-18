package bls

import (
	"encoding/hex"
	"testing"
)

func TestSecretKey(t *testing.T) {
	sk := NewSecretKeyByCSPRNG()
	s, err := sk.Sign([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Verify(sk.PublicKey(), []byte("hello")) != true {
		t.Error()
	}
	if s.Verify(sk.PublicKey(), []byte("world")) != false {
		t.Error()
	}
}

func TestVerify(t *testing.T) {
	// SOURCE: https://github.com/dfinity/agent-js/blob/5214dc1fc4b9b41f023a88b1228f04d2f2536987/packages/bls-verify/src/index.test.ts#L101
	publicKeyHex := "a7623a93cdb56c4d23d99c14216afaab3dfd6d4f9eb3db23d038280b6d5cb2caaee2a19dd92c9df7001dede23bf036bc0f33982dfb41e8fa9b8e96b5dc3e83d55ca4dd146c7eb2e8b6859cb5a5db815db86810b8d12cee1588b5dbf34a4dc9a5"
	publicKeyRaw, _ := hex.DecodeString(publicKeyHex)
	publicKey, err := PublicKeyFromBytes(publicKeyRaw)
	if err != nil {
		t.Fatal(err)
	}

	signatureHex := "b89e13a212c830586eaa9ad53946cd968718ebecc27eda849d9232673dcd4f440e8b5df39bf14a88048c15e16cbcaabe"
	signatureHexRaw, _ := hex.DecodeString(signatureHex)
	signature, err := SignatureFromBytes(signatureHexRaw)
	if err != nil {
		t.Fatal(err)
	}

	if signature.Verify(publicKey, []byte("bye")) {
		t.Error()
	}
	if !signature.Verify(publicKey, []byte("hello")) {
		t.Error()
	}
}

func TestVerify_hex(t *testing.T) {
	// SOURCE: https://github.com/dfinity/agent-js/blob/5214dc1fc4b9b41f023a88b1228f04d2f2536987/packages/bls-verify/src/index.test.ts#L101
	publicKeyHex := "a7623a93cdb56c4d23d99c14216afaab3dfd6d4f9eb3db23d038280b6d5cb2caaee2a19dd92c9df7001dede23bf036bc0f33982dfb41e8fa9b8e96b5dc3e83d55ca4dd146c7eb2e8b6859cb5a5db815db86810b8d12cee1588b5dbf34a4dc9a5"
	publicKey, err := PublicKeyFromHexString(publicKeyHex)
	if err != nil {
		t.Fatal(err)
	}

	signatureHex := "b89e13a212c830586eaa9ad53946cd968718ebecc27eda849d9232673dcd4f440e8b5df39bf14a88048c15e16cbcaabe"
	signature, err := SignatureFromHexString(signatureHex)
	if err != nil {
		t.Fatal(err)
	}

	if signature.Verify(publicKey, []byte("bye")) {
		t.Error()
	}
	if !signature.Verify(publicKey, []byte("hello")) {
		t.Error()
	}
}
