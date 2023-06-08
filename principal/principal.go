package principal

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"hash/crc32"
	"strings"
)

// AnonymousID is used for the anonymous caller. It can be used in call and query requests without a signature.
var AnonymousID = Principal{[]byte{0x04}}

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// Principal are generic identifiers for canisters, users and possibly other concepts in the future.
type Principal struct {
	Raw []byte
}

// Decode converts a textual representation into a principal.
func Decode(s string) (Principal, error) {
	p := strings.Split(s, "-")
	for i, c := range p {
		if len(c) > 5 {
			return Principal{}, fmt.Errorf("invalid length: %s", c)
		}
		if i != len(p)-1 && len(c) < 5 {
			return Principal{}, fmt.Errorf("invalid length: %s", c)
		}
	}
	b32, err := encoding.DecodeString(strings.ToUpper(strings.Join(p, "")))
	if err != nil {
		return Principal{}, err
	}
	if len(b32) < 4 {
		return Principal{}, fmt.Errorf("invalid length: %s", b32)
	}
	if crc32.ChecksumIEEE(b32[4:]) != binary.BigEndian.Uint32(b32[:4]) {
		return Principal{}, fmt.Errorf("invalid checksum: %s", b32)
	}
	return Principal{b32[4:]}, err
}

// NewSelfAuthenticating returns a self authenticating principal identifier based on the given public key.
func NewSelfAuthenticating(pub []byte) Principal {
	hash := sha256.Sum224(pub)
	return Principal{
		Raw: append(hash[:], 0x02),
	}
}

// Encode converts the principal to its textual representation.
func (p Principal) Encode() string {
	cs := make([]byte, 4)
	binary.BigEndian.PutUint32(cs, crc32.ChecksumIEEE(p.Raw))
	b32 := strings.ToLower(encoding.EncodeToString(append(cs, p.Raw...)))
	var str string
	for i, c := range b32 {
		if i != 0 && i%5 == 0 {
			str += "-"
		}
		str += string(c)
	}
	return str
}

// MarshalCBOR converts the principal to its CBOR representation.
func (p Principal) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(p.Raw)
}

// String implements the Stringer interface.
func (p Principal) String() string {
	return p.Encode()
}

// UnmarshalCBOR converts a CBOR representation into a principal.
func (p *Principal) UnmarshalCBOR(bytes []byte) error {
	return cbor.Unmarshal(bytes, &p.Raw)
}
