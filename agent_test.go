package agent_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/candid-go"
	"github.com/aviate-labs/principal-go"
)

func ExampleQuery() {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	var id identity.Identity = identity.NewEd25519Identity(publicKey, privateKey)
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	agent := agent.New(agent.AgentConfig{
		Identity: &id,
	})
	args, _ := candid.EncodeValue("record { account = \"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d\" }")
	fmt.Println(agent.Query(ledgerID, "account_balance_dfx", args))
	// Output:
	// (record { 5035232 = 0 : nat64 }) <nil>
}

func ExampleQuery_Anonymous() {
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	agent := agent.New(agent.AgentConfig{})
	args, _ := candid.EncodeValue("record { account = \"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d\" }")
	fmt.Println(agent.Query(ledgerID, "account_balance_dfx", args))
	// Output:
	// (record { 5035232 = 0 : nat64 }) <nil>
}
