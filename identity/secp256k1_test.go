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
	if !bytes.Equal(id.privateKey.Serialize(), id_.privateKey.Serialize()) {
		t.Error()
	}
	if !bytes.Equal(id.PublicKey(), id_.PublicKey()) {
		t.Error()
	}
}

func TestNewSecp256k1IdentityFromPEM(t *testing.T) {
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
	der, _ := hex.DecodeString("3056301006072a8648ce3d020106052b8104000a0342000480ef3bac9d68cf374cbc9c9943e180043a94c462ef8270274e57089d5dcdba1b8fbf83b7546ccc1b3781377776e9c4710fdf533ed9bcba8d8ebb32a14a2aba11")
	id, err := NewSecp256k1IdentityFromPEM([]byte(pem))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id.PublicKey(), der) {
		t.Fatal("public key mismatch")
	}
}

func TestSecp256k1Identity_Sign(t *testing.T) {
	id, err := NewRandomSecp256k1Identity()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello")
	if !id.Verify(data, id.Sign(data)) {
		t.Error()
	}
}
