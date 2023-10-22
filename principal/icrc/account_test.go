package icrc_test

import (
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/aviate-labs/agent-go/principal/icrc"
	"testing"
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

func TestAccountIdentifier(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := icrc.Account{
			Owner:      principal.Principal{},
			SubAccount: &[32]byte{byte(i)},
		}
		accountID, err := icrc.Decode(a.String())
		if err != nil {
			t.Error(err)
		}
		if accountID.String() != a.String() {
			t.Errorf("expected %s, got %s", a.String(), accountID.String())
		}
	}
}

func TestVectors(t *testing.T) {
	for _, test := range []struct {
		account string
		err     bool
	}{
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae"},
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae-q6bn32y.", err: true},
		{account: "k2t6j2nvnp4zjm3-25dtz6xhaac7boj5gayfoj3xs-i43lp-teztq-6ae", err: true},
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae-6cc627i.1"},
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae-6cc627i.01", err: true},
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae.1", err: true},
		{account: "k2t6j-2nvnp-4zjm3-25dtz-6xhaa-c7boj-5gayf-oj3xs-i43lp-teztq-6ae-dfxgiyy.102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"},
	} {
		a, err := icrc.Decode(test.account)
		if err != nil {
			if !test.err {
				t.Error(err)
			}
			continue
		}
		if a.String() != test.account {
			t.Errorf("expected %s, got %s", test.account, a.String())
		}
	}
}
