package verify

import (
	"bytes"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
)

var (
	CanisterSigPublicKeyDERObjectID = []byte{
		0x30, 0x0C, 0x06, 0x0A, 0x2B, 0x06, 0x01,
		0x04, 0x01, 0x83, 0xB8, 0x43, 0x01, 0x02,
	}
	CanisterSigPublicKeyPrefixLength = 19
)

type CanisterSigPublicKey struct {
	CanisterID principal.Principal
	Seed       []byte
}

func CanisterSigPublicKeyFromDER(der []byte) (*CanisterSigPublicKey, error) {
	if len(der) < 21 {
		return nil, fmt.Errorf("DER data is too short")
	}
	if !bytes.Equal(der[2:len(CanisterSigPublicKeyDERObjectID)+2], CanisterSigPublicKeyDERObjectID) {
		return nil, fmt.Errorf("DER data does not match object ID")
	}
	canisterIDLength := int(der[CanisterSigPublicKeyPrefixLength])
	if len(der) < CanisterSigPublicKeyPrefixLength+canisterIDLength {
		return nil, fmt.Errorf("DER data is too short")
	}
	offset := CanisterSigPublicKeyPrefixLength + 1
	rawCanisterID := der[offset : offset+canisterIDLength]
	offset += canisterIDLength
	return &CanisterSigPublicKey{
		CanisterID: principal.Principal{Raw: rawCanisterID},
		Seed:       der[offset:],
	}, nil
}

func (s *CanisterSigPublicKey) DER() []byte {
	raw := s.Raw()
	var der bytes.Buffer
	der.WriteByte(0x30)
	der.WriteByte(17 + byte(len(raw)))
	der.Write(CanisterSigPublicKeyDERObjectID)
	der.WriteByte(0x03)
	der.WriteByte(1 + byte(len(raw)))
	der.WriteByte(0x00)
	der.Write(raw)
	return der.Bytes()
}

func (s *CanisterSigPublicKey) Raw() []byte {
	var raw bytes.Buffer
	raw.WriteByte(byte(len(s.CanisterID.Raw)))
	raw.Write(s.CanisterID.Raw)
	raw.Write(s.Seed)
	return raw.Bytes()
}
