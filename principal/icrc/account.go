package icrc

import (
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"hash/crc32"
	"strings"
)

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

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

func Decode(s string) (Account, error) {
	a := strings.Split(s, ".")
	if len(a) == 1 {
		owner, err := principal.Decode(s)
		if err != nil {
			return Account{}, err
		}
		return Account{
			Owner:      owner,
			SubAccount: nil,
		}, nil
	}
	if len(a) != 2 {
		return Account{}, fmt.Errorf("invalid account identifier: %s", s)
	}
	p := strings.Split(a[0], "-")
	b32crc := strings.ToUpper(p[len(p)-1])
	owner, err := principal.Decode(strings.Join(p[:len(p)-1], "-"))
	if err != nil {
		return Account{}, err
	}
	if len(a[1]) == 0 || a[1][0] == '0' {
		return Account{}, fmt.Errorf("invalid sub account: %s", a[1])
	}
	if len(a[1])%2 == 1 {
		// Add leading zero if necessary.
		a[1] = "0" + a[1]
	}
	subAccount, err := hex.DecodeString(a[1])
	if err != nil {
		return Account{}, err
	}
	for len(subAccount) < 32 {
		subAccount = append([]byte{0}, subAccount...)
	}
	cs, err := encoding.DecodeString(b32crc)
	if err != nil {
		return Account{}, err
	}
	if len(cs) != 4 {
		return Account{}, fmt.Errorf("invalid checksum size: %d", len(cs))
	}
	if crc32.ChecksumIEEE(append(owner.Raw, subAccount...)) != binary.BigEndian.Uint32(cs) {
		return Account{}, fmt.Errorf("invalid checksum: %s", string(cs))
	}
	var subAccount32 [32]byte
	copy(subAccount32[:], subAccount)
	return Account{
		Owner:      owner,
		SubAccount: &subAccount32,
	}, nil
}

func (a Account) Encode() string {
	if a.SubAccount == nil {
		return a.Owner.String()
	}
	if *a.SubAccount == [32]byte{} {
		return a.Owner.String()
	}
	cs := make([]byte, 4)
	binary.BigEndian.PutUint32(cs, crc32.ChecksumIEEE(append(a.Owner.Raw, a.SubAccount[:]...)))
	b32cs := strings.ToLower(encoding.EncodeToString(cs))
	return a.Owner.String() + "-" + b32cs + "." + trimLeadingZeros(hex.EncodeToString(a.SubAccount[:]))
}

func (a Account) String() string {
	return a.Encode()
}
