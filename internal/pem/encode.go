package pem

import (
	"crypto/elliptic"
	"encoding/asn1"
	"encoding/pem"

	"github.com/aviate-labs/secp256k1"
)

var curve = secp256k1.S256()

func Encode(key *secp256k1.PrivateKey) ([]byte, error) {
	p, err := asn1.Marshal(asn1.ObjectIdentifier{
		1, 3, 132, 0, 10,
	})
	if err != nil {
		return nil, err
	}
	k, err := encodePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return append(
		pem.EncodeToMemory(&pem.Block{
			Type:  "PARAMETERS",
			Bytes: p,
		}),
		pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: k,
		})...,
	), nil
}

func encodePrivateKey(key *secp256k1.PrivateKey) ([]byte, error) {
	p := key.PubKey().ToECDSA()
	return asn1.Marshal(asn1PrivateKey{
		Version: 1,
		MetaData: []asn1.ObjectIdentifier{
			{ // http://oid-info.com/get/1.3.132.0.10
				1, 3, 132, 0, 10,
			},
		},
		PrivateKey: asn1.BitString{
			Bytes: key.D.Bytes(),
		},
		PublicKey: asn1.BitString{
			Bytes: elliptic.Marshal(curve, p.X, p.Y),
		},
	})
}

func encodePublicKey(key *secp256k1.PublicKey) ([]byte, error) {
	p := key.ToECDSA()
	return asn1.Marshal(asn1PrivateKey{
		Version: 1,
		MetaData: []asn1.ObjectIdentifier{
			{ // http://www.oid-info.com/get/1.2.840.10045.2.1
				1, 2, 840, 10045, 2, 1,
			},
			{ // http://oid-info.com/get/1.3.132.0.10
				1, 3, 132, 0, 10,
			},
		},
		PublicKey: asn1.BitString{
			Bytes: elliptic.Marshal(curve, p.X, p.Y),
		},
	})
}

type asn1PrivateKey struct {
	Version    int
	PrivateKey asn1.BitString
	PublicKey  asn1.BitString
	MetaData   []asn1.ObjectIdentifier
}

type asn1PublicKey struct {
	PublicKey asn1.BitString
	Metadata  []asn1.ObjectIdentifier
}
