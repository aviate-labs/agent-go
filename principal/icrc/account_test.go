package icrc_test

import (
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/agent-go/principal/icrc"
)

func ExampleAccount() {
	p, _ := principal.Decode("k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae")
	s, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	var s32 [32]byte
	copy(s32[:], s)
	fmt.Println(icrc.Account{
		Owner:      p,
		SubAccount: &s32,
	}.String())
	// Output:
	// k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae-dfxgiyy.102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20
}
