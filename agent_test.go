package agent_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
)

func Example_anonymous_query() {
	a, _ := agent.New(agent.DefaultConfig)
	type Account struct {
		Account string `ic:"account"`
	}
	var balance struct {
		E8S uint64 `ic:"e8s"`
	}
	_ = a.Query(ic.LEDGER_PRINCIPAL, "account_balance_dfx", []any{
		Account{"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d"},
	}, []any{&balance})
	fmt.Println(balance.E8S)
	// Output:
	// 0
}

func Example_json() {
	raw := `{"e8s":1}`
	var balance struct {
		// Tags can be combined with json tags.
		E8S uint64 `ic:"e8s" json:"e8s"`
	}
	_ = json.Unmarshal([]byte(raw), &balance)
	fmt.Println(balance.E8S)

	a, _ := agent.New(agent.Config{})
	_ = a.Query(ic.LEDGER_PRINCIPAL, "account_balance_dfx", []any{struct {
		Account string `json:"account"`
	}{
		Account: "9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d",
	}}, []any{&balance}) // Repurposing the balance struct.
	rawJSON, _ := json.Marshal(balance)
	fmt.Println(string(rawJSON))
	// Output:
	// 1
	// {"e8s":0}
}

func Example_query() {
	publicKey, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	id, _ := identity.NewEd25519Identity(publicKey, privateKey)
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	a, _ := agent.New(agent.Config{Identity: id})
	var balance struct {
		E8S uint64 `ic:"e8s"`
	}
	_ = a.Query(ledgerID, "account_balance_dfx", []any{map[string]any{
		"account": "9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d",
	}}, []any{&balance})
	fmt.Println(balance.E8S)
	// Output:
	// 0
}
