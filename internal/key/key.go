package key

import (
	"fmt"

	"github.com/aviate-labs/bip32"
	"github.com/aviate-labs/bip39"
	"github.com/aviate-labs/secp256k1"
)

var (
	curve = secp256k1.S256()
)

func DeriveKeyPair(key bip32.Key, index uint32) (*secp256k1.PrivateKey, *secp256k1.PublicKey, error) {
	child0, err := key.NewChildKey(0)
	if err != nil {
		return nil, nil, err
	}
	child0i, err := child0.NewChildKey(index)
	if err != nil {
		return nil, nil, err
	}
	return Keys(child0i)
}

func Keys(key bip32.Key) (*secp256k1.PrivateKey, *secp256k1.PublicKey, error) {
	if key.IsPublic {
		return nil, nil, fmt.Errorf("can not create private key from a public key")
	}
	s, err := key.Serialize()
	if err != nil {
		return nil, nil, err
	}
	priv, pub := secp256k1.PrivKeyFromBytes(curve, s)
	return priv, pub, nil
}

func New(words bip39.Mnemonic, password string) (bip32.Key, error) {
	seed := bip39.NewSeed(words, password)
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return bip32.Key{}, err
	}
	return master.NewChildKey(0)
}

func RandomPrivateKey() (*secp256k1.PrivateKey, error) {
	e, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}
	m, err := bip39.English.NewMnemonic(e)
	if err != nil {
		return nil, err
	}
	n, err := New(m, "")
	if err != nil {
		return nil, err
	}
	priv, _, err := Keys(n)
	return priv, err
}
