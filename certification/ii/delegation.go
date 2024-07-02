package ii

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

type BEHexUint64 uint64

func (b *BEHexUint64) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	bb, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*b = BEHexUint64(binary.BigEndian.Uint64(bb))
	return nil
}

type Delegation struct {
	PublicKey  HexString   `json:"pubkey"`
	Expiration BEHexUint64 `json:"expiration"`
	Targets    []HexString `json:"targets"`
}

func (d Delegation) SignatureMessage() ([]byte, error) {
	kv := []certification.KeyValuePair{
		{Key: "pubkey", Value: []byte(d.PublicKey)},
		{Key: "expiration", Value: uint64(d.Expiration)},
	}
	ts := make([]any, len(d.Targets))
	for i, target := range d.Targets {
		ts[i] = []byte(target)
	}
	if 0 < len(ts) {
		kv = append(kv, certification.KeyValuePair{Key: "targets", Value: ts})
	}
	hash, err := certification.RepresentationIndependentHash(kv)
	if err != nil {
		return nil, err
	}
	return append([]byte("\x1aic-request-auth-delegation"), hash[:]...), nil
}

type DelegationChain struct {
	Delegations []SignedDelegation `json:"delegations"`
	PublicKey   HexString          `json:"publicKey"`
}

func (d DelegationChain) VerifyChallenge(
	challenge []byte,
	currentTimeNS uint64,
	canisterID principal.Principal,
	rootPublicKey []byte,
) error {
	if len(d.Delegations) != 1 {
		return fmt.Errorf("expected exactly one delegation")
	}
	signedDelegation := d.Delegations[0]
	delegation := signedDelegation.Delegation
	if !bytes.Equal(challenge, []byte(delegation.PublicKey)) {
		return fmt.Errorf("invalid challenge")
	}
	canisterSig, err := CanisterSigPublicKeyFromDER([]byte(d.PublicKey))
	if err != nil {
		return err
	}
	if !bytes.Equal(canisterSig.CanisterID.Raw, canisterID.Raw) {
		return fmt.Errorf("invalid canister ID")
	}
	if uint64(delegation.Expiration) < currentTimeNS {
		return fmt.Errorf("delegation expired")
	}

	sig := []byte(signedDelegation.Signature)
	message, err := delegation.SignatureMessage()
	if err != nil {
		return err
	}
	var wrapper struct {
		Certificate []byte            `cbor:"certificate"`
		Tree        hashtree.HashTree `cbor:"tree"`
	}
	if err := cbor.Unmarshal(sig, &wrapper); err != nil {
		return err
	}
	var certificate certification.Certificate
	if err := cbor.Unmarshal(wrapper.Certificate, &certificate); err != nil {
		return err
	}
	tree := wrapper.Tree.Digest()
	if err := certification.VerifyCertifiedData(
		certificate,
		canisterSig.CanisterID,
		rootPublicKey,
		tree[:],
	); err != nil {
		return err
	}
	seed := sha256.Sum256(canisterSig.Seed)
	msg := sha256.Sum256(message)
	if _, err := wrapper.Tree.Lookup(hashtree.Label("sig"), seed[:], msg[:]); err != nil {
		return err
	}
	return nil
}

type HexString string

func (h *HexString) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	*h = HexString(b)
	return nil
}

type SignedDelegation struct {
	Delegation Delegation `json:"delegation"`
	Signature  HexString  `json:"signature"`
}
