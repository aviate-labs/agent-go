package pem

import (
	"encoding/asn1"
	"encoding/pem"
	"fmt"

	"github.com/aviate-labs/secp256k1"
)

func Decode(data []byte) (*secp256k1.PrivateKey, error) {
	var block *pem.Block
	block, data = pem.Decode(data)
	if block == nil || block.Type != "PARAMETERS" {
		return nil, fmt.Errorf("no parameter block")
	}
	var params asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(block.Bytes, &params); err != nil {
		return nil, err
	}
	if !params.Equal(asn1.ObjectIdentifier{
		1, 3, 132, 0, 10,
	}) {
		return nil, fmt.Errorf("invalid params")
	}
	block, _ = pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("no private key block")
	}
	var keyBlock asn1PrivateKey
	if _, err := asn1.Unmarshal(block.Bytes, &keyBlock); err != nil {
		return nil, err
	}
	// TODO: validate version and metadata
	privateKey, _ := secp256k1.PrivKeyFromBytes(curve, keyBlock.PrivateKey.Bytes)
	return privateKey, nil
}
