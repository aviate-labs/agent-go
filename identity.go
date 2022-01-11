package agent

import "github.com/aviate-labs/secp256k1"

type Identity struct {
	privateKey *secp256k1.PrivateKey
}

func (i *Identity) Sign(message []byte) ([]byte, error) {
	signature, err := i.privateKey.Sign(message)
	if err != nil {
		return nil, err
	}
	var b [64]byte
	r := signature.R.Bytes()
	s := signature.S.Bytes()
	copy(b[(32-len(r)):], r)
	copy(b[(64-len(s)):], s)
	return b[:], nil
}
