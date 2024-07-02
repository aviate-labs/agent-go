package certification

import (
	"bytes"
	"crypto/ed25519"
	"encoding/asn1"
	"fmt"
	"github.com/aviate-labs/agent-go/certification/bls"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/leb128"
	"slices"
	"time"

	"github.com/fxamacker/cbor/v2"
)

func PublicBLSKeyFromDER(der []byte) (*bls.PublicKey, error) {
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(der, &seq); err != nil {
		return nil, err
	}
	if seq.Tag != asn1.TagSequence {
		return nil, fmt.Errorf("invalid tag: %d", seq.Tag)
	}
	var idSeq asn1.RawValue
	rest, err := asn1.Unmarshal(seq.Bytes, &idSeq)
	if err != nil {
		return nil, err
	}
	var bs asn1.BitString
	if _, err := asn1.Unmarshal(rest, &bs); err != nil {
		return nil, err
	}
	if bs.BitLength != 96*8 {
		return nil, fmt.Errorf("invalid bit string length: %d", bs.BitLength)
	}
	var algoId asn1.ObjectIdentifier
	seqRest, err := asn1.Unmarshal(idSeq.Bytes, &algoId)
	if err != nil {
		return nil, err
	}
	if !algoId.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 44668, 5, 3, 1, 2, 1}) {
		return nil, fmt.Errorf("invalid algorithm identifier: %v", algoId)
	}
	var curveID asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(seqRest, &curveID); err != nil {
		return nil, err
	}
	if !curveID.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 44668, 5, 3, 2, 1}) {
		return nil, fmt.Errorf("invalid curve identifier: %v", curveID)
	}
	return bls.PublicKeyFromBytes(bs.Bytes)
}

func PublicBLSKeyToDER(publicKey []byte) ([]byte, error) {
	if len(publicKey) != 96 {
		return nil, fmt.Errorf("invalid public key length: %d", len(publicKey))
	}
	return asn1.Marshal([]any{
		[]any{
			asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 44668, 5, 3, 1, 2, 1}, // algorithm identifier
			asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 44668, 5, 3, 2, 1},    // curve identifier
		},
		asn1.BitString{
			Bytes:     publicKey,
			BitLength: len(publicKey) * 8,
		},
	})
}

func PublicED25519KeyFromDER(der []byte) (*ed25519.PublicKey, error) {
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(der, &seq); err != nil {
		return nil, err
	}
	if seq.Tag != asn1.TagSequence {
		return nil, fmt.Errorf("invalid tag: %d", seq.Tag)
	}
	var idSeq asn1.RawValue
	rest, err := asn1.Unmarshal(seq.Bytes, &idSeq)
	if err != nil {
		return nil, err
	}
	var bs asn1.BitString
	if _, err := asn1.Unmarshal(rest, &bs); err != nil {
		return nil, err
	}
	var algoId asn1.ObjectIdentifier
	if _, err := asn1.Unmarshal(idSeq.Bytes, &algoId); err != nil {
		return nil, err
	}
	if !algoId.Equal(asn1.ObjectIdentifier{1, 3, 101, 112}) {
		return nil, fmt.Errorf("invalid algorithm identifier: %v", algoId)
	}
	publicKey := ed25519.PublicKey(bs.Bytes)
	return &publicKey, nil
}
func VerifyCertificate(
	certificate Certificate,
	canisterID principal.Principal,
	rootPublicKey []byte,
) error {
	publicKey, err := PublicBLSKeyFromDER(rootPublicKey)
	if err != nil {
		return err
	}
	key := publicKey
	if certificate.Delegation != nil {
		delegation := certificate.Delegation
		k, err := verifyDelegationCertificate(
			delegation,
			publicKey,
			canisterID,
		)
		if err != nil {
			return err
		}
		key = k
	}
	return verifyCertificateSignature(certificate, key)
}

func VerifyCertifiedData(
	certificate Certificate,
	canisterID principal.Principal,
	rootPublicKey []byte,
	certifiedData []byte,
) error {
	if err := VerifyCertificate(certificate, canisterID, rootPublicKey); err != nil {
		return err
	}
	certificateCertifiedData, err := certificate.Tree.Lookup(
		hashtree.Label("canister"),
		canisterID.Raw,
		hashtree.Label("certified_data"),
	)
	if err != nil {
		return err
	}
	if !bytes.Equal(certificateCertifiedData, certifiedData) {
		return fmt.Errorf("certified data does not match: %x != %x", certificateCertifiedData, certifiedData)
	}
	return nil
}

func VerifySubnetCertificate(
	certificate Certificate,
	subnetID principal.Principal,
	rootPublicKey []byte,
) error {
	publicKey, err := PublicBLSKeyFromDER(rootPublicKey)
	if err != nil {
		return err
	}
	return verifySubnetCertificate(certificate, subnetID, publicKey)
}

func verifyCertificateSignature(certificate Certificate, publicKey *bls.PublicKey) error {
	rootHash := certificate.Tree.Digest()
	message := append(hashtree.DomainSeparator("ic-state-root"), rootHash[:]...)
	signature, err := bls.SignatureFromBytes(certificate.Signature)
	if err != nil {
		return err
	}
	if !signature.VerifyByte(publicKey, message) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func verifyDelegationCertificate(
	delegation *Delegation,
	rootPublicKey *bls.PublicKey,
	canisterID principal.Principal,
) (*bls.PublicKey, error) {
	if delegation.Certificate.Delegation != nil {
		return nil, fmt.Errorf("multiple delegations are not supported")
	}
	if err := verifyCertificateSignature(delegation.Certificate, rootPublicKey); err != nil {
		return nil, err
	}

	rawRanges, err := delegation.Certificate.Tree.Lookup(
		hashtree.Label("subnet"),
		delegation.SubnetId.Raw,
		hashtree.Label("canister_ranges"),
	)
	if err != nil {
		return nil, err
	}
	var canisterRanges CanisterRanges
	if err := cbor.Unmarshal(rawRanges, &canisterRanges); err != nil {
		return nil, err
	}
	if !canisterRanges.InRange(canisterID) {
		return nil, fmt.Errorf("canister %s is not in range", canisterID)
	}

	rawPublicKey, err := delegation.Certificate.Tree.Lookup(
		hashtree.Label("subnet"),
		delegation.SubnetId.Raw,
		hashtree.Label("public_key"),
	)
	if err != nil {
		return nil, err
	}
	return PublicBLSKeyFromDER(rawPublicKey)
}

func verifySubnetCertificate(
	certificate Certificate,
	subnetID principal.Principal,
	rootPublicKey *bls.PublicKey,
) error {
	key := rootPublicKey
	if certificate.Delegation != nil {
		delegation := certificate.Delegation
		k, err := verifySubnetDelegationCertificate(
			delegation,
			subnetID,
			rootPublicKey,
		)
		if err != nil {
			return err
		}
		key = k
	}
	return verifyCertificateSignature(certificate, key)
}

func verifySubnetDelegationCertificate(
	delegation *Delegation,
	subnetID principal.Principal,
	rootPublicKey *bls.PublicKey,
) (*bls.PublicKey, error) {
	if delegation.Certificate.Delegation != nil {
		return nil, fmt.Errorf("multiple delegations are not supported")
	}
	if err := verifySubnetCertificate(delegation.Certificate, subnetID, rootPublicKey); err != nil {
		return nil, err
	}

	rawPublicKey, err := delegation.Certificate.Tree.Lookup(
		hashtree.Label("subnet"),
		subnetID.Raw,
		hashtree.Label("public_key"),
	)
	if err != nil {
		return nil, err
	}
	return PublicBLSKeyFromDER(rawPublicKey)
}

type CanisterRange struct {
	From principal.Principal
	To   principal.Principal
}

func (c *CanisterRange) UnmarshalCBOR(bytes []byte) error {
	var raw [][]byte
	if err := cbor.Unmarshal(bytes, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("unexpected length: %d", len(raw))
	}
	c.From = principal.Principal{Raw: raw[0]}
	c.To = principal.Principal{Raw: raw[1]}
	return nil
}

type CanisterRanges []CanisterRange

func (c CanisterRanges) InRange(canisterID principal.Principal) bool {
	for _, r := range c {
		if slices.Compare(r.From.Raw, canisterID.Raw) <= 0 && slices.Compare(canisterID.Raw, r.To.Raw) <= 0 {
			return true
		}
	}
	return false
}

// Certificate is a certificate gets returned by the IC.
type Certificate struct {
	// Tree is the certificate tree.
	Tree hashtree.HashTree `cbor:"tree"`
	// Signature is the signature of the certificate tree.
	Signature []byte `cbor:"signature"`
	// Delegation is the delegation of the certificate.
	Delegation *Delegation `cbor:"delegation"`
}

// VerifyTime verifies the time of a certificate.
func (c Certificate) VerifyTime(ingressExpiry time.Duration) error {
	rawTime, err := c.Tree.Lookup(hashtree.Label("time"))
	if err != nil {
		return err
	}
	t, err := leb128.DecodeUnsigned(bytes.NewReader(rawTime))
	if err != nil {
		return err
	}
	if int64(ingressExpiry) < time.Now().UnixNano()-t.Int64() {
		return fmt.Errorf("certificate outdated, exceeds ingress expiry")
	}
	return nil
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
	var m map[string][]byte
	if err := cbor.Unmarshal(bytes, &m); err != nil {
		return err
	}
	for k, v := range m {
		switch k {
		case "subnet_id":
			d.SubnetId = principal.Principal{
				Raw: v,
			}
		case "certificate":
			if err := cbor.Unmarshal(v, &d.Certificate); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}
	return nil
}
