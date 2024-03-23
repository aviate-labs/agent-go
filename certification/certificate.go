package certification

import (
	"fmt"
	"slices"

	"github.com/aviate-labs/agent-go/certification/bls"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"

	"github.com/fxamacker/cbor/v2"
)

// Cert is a certificate gets returned by the IC.
type Cert struct {
	// Tree is the certificate tree.
	Tree hashtree.HashTree `cbor:"tree"`
	// Signature is the signature of the certificate tree.
	Signature []byte `cbor:"signature"`
	// Delegation is the delegation of the certificate.
	Delegation *Delegation `cbor:"delegation"`
}

// Certificate is a certificate that gets returned by the IC and can be used to verify the state root based on the root
// key and canister ID.
type Certificate struct {
	Cert       Cert
	RootKey    []byte
	CanisterID principal.Principal
}

// New creates a new certificate.
func New(canisterID principal.Principal, rootKey []byte, certificate []byte) (*Certificate, error) {
	var cert Cert
	if err := cbor.Unmarshal(certificate, &cert); err != nil {
		return nil, err
	}
	return &Certificate{
		Cert:       cert,
		RootKey:    rootKey,
		CanisterID: canisterID,
	}, nil
}

// Verify verifies the certificate.
func (c Certificate) Verify() error {
	signature, err := bls.SignatureFromBytes(c.Cert.Signature)
	if err != nil {
		return err
	}
	publicKey, err := c.getPublicKey()
	if err != nil {
		return err
	}
	rootHash := c.Cert.Tree.Digest()
	message := append(hashtree.DomainSeparator("ic-state-root"), rootHash[:]...)
	if !signature.Verify(publicKey, string(message)) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

// getPublicKey checks the delegation and returns the public key.
func (c Certificate) getPublicKey() (*bls.PublicKey, error) {
	if c.Cert.Delegation == nil {
		return bls.PublicKeyFromBytes(c.RootKey)
	}

	cert := c.Cert.Delegation
	canisterRangesResult := cert.Certificate.Cert.Tree.Lookup(
		hashtree.Label("subnet"), cert.SubnetId.Raw, hashtree.Label("canister_ranges"),
	)
	if canisterRangesResult.Found() != nil {
		return nil, fmt.Errorf("no canister ranges found for subnet %s", cert.SubnetId)
	}
	var rawRanges [][][]byte
	if err := cbor.Unmarshal(canisterRangesResult.Value, &rawRanges); err != nil {
		return nil, err
	}

	var inRange bool
	for _, pair := range rawRanges {
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid range: %v", pair)
		}
		if slices.Compare(pair[0], c.CanisterID.Raw) <= 0 && slices.Compare(c.CanisterID.Raw, pair[1]) <= 0 {
			inRange = true
			break
		}
	}
	if !inRange {
		return nil, fmt.Errorf("canister %s is not in range", c.CanisterID)
	}

	publicKeyResult := cert.Certificate.Cert.Tree.Lookup(
		hashtree.Label("subnet"), cert.SubnetId.Raw, hashtree.Label("public_key"),
	)
	if publicKeyResult.Found() != nil {
		return nil, fmt.Errorf("no public key found for subnet %s", cert.SubnetId)
	}

	publicKey := publicKeyResult.Value
	if len(publicKey) != len(derPrefix)+96 {
		return nil, fmt.Errorf("invalid public key length: %d", len(publicKey))
	}

	if slices.Compare(publicKey[:len(derPrefix)], derPrefix) != 0 {
		return nil, fmt.Errorf("invalid public key prefix: %s", publicKey[:len(derPrefix)])
	}

	return bls.PublicKeyFromBytes(publicKey[len(derPrefix):])
}

// Delegation is a delegation of a certificate.
type Delegation struct {
	// SubnetId is the subnet ID of the delegation.
	SubnetId principal.Principal `cbor:"subnet_id"`
	// The nested certificate typically does not itself again contain a
	// delegation, although there is no reason why agents should enforce that
	// property.
	Certificate Certificate `cbor:"certificate"`
}

// UnmarshalCBOR unmarshals a delegation.
func (d *Delegation) UnmarshalCBOR(bytes []byte) error {
	var m map[string]any
	if err := cbor.Unmarshal(bytes, &m); err != nil {
		return err
	}
	for k, v := range m {
		switch k {
		case "subnet_id":
			d.SubnetId = principal.Principal{
				Raw: v.([]byte),
			}
		case "certificate":
			if err := cbor.Unmarshal(v.([]byte), &d.Certificate.Cert); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}
	return nil
}
