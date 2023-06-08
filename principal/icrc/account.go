package icrc

import (
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"github.com/aviate-labs/agent-go/principal"
	"hash/crc32"
	"strings"
)

func trimLeadingZeros(str string) string {
	for str[0] == '0' {
		str = str[1:]
	}
	return str
}

type Account struct {
	Owner      principal.Principal
	SubAccount *[32]byte
}

func (a Account) String() string {
	if a.SubAccount == nil {
		return a.Owner.String()
	}
	if *a.SubAccount == [32]byte{} {
		return a.Owner.String()
	}
	cs := make([]byte, 4)
	binary.BigEndian.PutUint32(cs, crc32.ChecksumIEEE(append(a.Owner.Raw, a.SubAccount[:]...)))
	b32cs := strings.ToLower(removePadding(base32.StdEncoding.EncodeToString(cs)))
	return a.Owner.String() + "-" + b32cs + "." + trimLeadingZeros(hex.EncodeToString(a.SubAccount[:]))
}

func removePadding(str string) string {
	for strings.HasSuffix(str, "=") {
		str = str[:len(str)-1]
	}
	return str
}
