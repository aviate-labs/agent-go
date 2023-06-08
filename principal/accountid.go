package principal

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
)

var (
	// DefaultSubAccount is the default sub-account represented by the empty byte array.
	DefaultSubAccount [32]byte
)

type (
	// AccountIdentifier is a unique identifier for an account.
	AccountIdentifier [28]byte
	// SubAccount is the sub-account identifier for an account.
	SubAccount [32]byte
)

// NewAccountID returns the account identifier corresponding with the given sub-account.
func NewAccountID(p Principal, subAccount [32]byte) AccountIdentifier {
	h := sha256.New224()
	h.Write([]byte("\x0Aaccount-id"))
	h.Write(p.Raw)
	h.Write(subAccount[:])
	bs := h.Sum(nil)

	var accountId [28]byte
	copy(accountId[:], bs)
	return accountId
}

// Bytes returns the bytes of the account identifier, including the checksum.
func (id AccountIdentifier) Bytes() []byte {
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, crc32.ChecksumIEEE(id[:]))
	return append(crc, id[:]...)
}

// DecodeAccountID decodes the given string into an account identifier.
func DecodeAccountID(s string) (AccountIdentifier, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return AccountIdentifier{}, err
	}
	if len(b) != 32 {
		return AccountIdentifier{}, fmt.Errorf("invalid length: %d", len(b))
	}
	crc := binary.BigEndian.Uint32(b[:4])
	if crc != crc32.ChecksumIEEE(b[4:]) {
		return AccountIdentifier{}, fmt.Errorf("invalid checksum: %d", crc)
	}
	var id AccountIdentifier
	copy(id[:], b[4:])
	return id, nil
}

// Encode returns the hexadecimal representation of the account identifier.
func (id AccountIdentifier) Encode() string {
	return hex.EncodeToString(id.Bytes())
}

// String returns the hexadecimal representation of the account identifier.
func (id AccountIdentifier) String() string {
	return id.Encode()
}
