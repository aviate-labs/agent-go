package principal

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
)

var (
	DefaultSubAccount [32]byte
)

type (
	AccountIdentifier [28]byte
	SubAccount        [32]byte
)

// Returns the account identifier corresponding with the given sub-account.
func (p Principal) AccountIdentifier(subAccount [32]byte) AccountIdentifier {
	h := sha256.New224()
	h.Write([]byte("\x0Aaccount-id"))
	h.Write(p.Raw)
	h.Write(subAccount[:])
	bs := h.Sum(nil)

	var accountId [28]byte
	copy(accountId[:], bs)
	return accountId
}

func (id AccountIdentifier) String() string {
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, crc32.ChecksumIEEE(id[:]))
	return hex.EncodeToString(append(crc, id[:]...))
}
