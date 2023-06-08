package principal

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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

// FromPrincipal returns the account identifier corresponding with the given sub-account.
func FromPrincipal(p Principal, subAccount [32]byte) AccountIdentifier {
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

// String returns the hexadecimal representation of the account identifier.
func (id AccountIdentifier) String() string {
	return hex.EncodeToString(id.Bytes())
}
