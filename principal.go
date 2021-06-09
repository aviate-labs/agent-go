package agent

import (
	"encoding/base32"
	"hash/crc32"
	"strings"
)

// AnonymousID is used for the anonymous caller. It can be used in call and query requests without a signature.
var AnonymousID = Principal([]byte{0x04})

// DecodePrincipal converts a textual representation into a principal.
func DecodePrincipal(s string) (Principal, error) {
	s = strings.ReplaceAll(s, "-", "")
	if i := len(s) % 8; i != 0 {
		s += strings.Repeat("=", 8-i)
	}
	s = strings.ToUpper(s)
	b32, err := base32.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return b32[4:], err
}

// Principal are generic identifiers for canisters, users and possibly other concepts in the future.
// More info: https://sdk.dfinity.org/docs/interface-spec/index.html#principal
type Principal []byte

// Encode converts the principal to its textual representation.
func (p Principal) Encode() string {
	h := crc32.NewIEEE()
	h.Write(p)
	b32 := base32.StdEncoding.EncodeToString(append(h.Sum(nil), p...))
	b32 = strings.TrimRight(b32, "=")
	b32 = strings.ToLower(b32)
	var str string
	for i, c := range b32 {
		if i != 0 && i%5 == 0 {
			str += "-"
		}
		str += string(c)
	}
	return str
}
