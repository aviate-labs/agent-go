package agent_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
)

func Example_anonymous_query() {
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	a := agent.New(agent.Config{})
	args, err := candid.EncodeValueString("record { account = \"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d\" }")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a.QueryString(ledgerID, "account_balance_dfx", args))
	// Output:
	// (record { 5035232 = 0 : nat64 }) <nil>
}

func Example_query() {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	id, _ := identity.NewEd25519Identity(publicKey, privateKey)
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	a := agent.New(agent.Config{
		Identity: id,
	})
	args, err := candid.EncodeValueString("record { account = \"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d\" }")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a.QueryString(ledgerID, "account_balance_dfx", args))
	// Output:
	// (record { 5035232 = 0 : nat64 }) <nil>
}
