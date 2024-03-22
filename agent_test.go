package agent_test

import (
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/mgmt"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
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

func Example_query_ed25519() {
	id, _ := identity.NewRandomEd25519Identity()
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

func Example_query_prime256v1() {
	id, _ := identity.NewRandomPrime256v1Identity()
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

func Example_query_secp256k1() {
	id, _ := identity.NewRandomSecp256k1Identity()
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

func TestAgent_Call(t *testing.T) {
	a, err := mgmt.NewAgent(agent.Config{
		Logger: &testLogger{},
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := a.BitcoinGetBalanceQuery(mgmt.BitcoinGetBalanceQueryArgs{
		Address: "bc1qruu3xmfrt4nzkxax3lpxfmjega87jr3vqcwjn9",
		Network: mgmt.BitcoinNetwork{
			Mainnet: new(idl.Null),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}

func TestICPLedger_queryBlocks(t *testing.T) {
	a, err := icpledger.NewAgent(ic.LEDGER_PRINCIPAL, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.QueryBlocks(icpledger.GetBlocksArgs{
		Start:  1,  // Start from the first block.
		Length: 10, // Get the last 10 blocks.
	}); err != nil {
		t.Fatal(err)
	}
}

type testLogger struct{}

func (l *testLogger) Printf(format string, v ...interface{}) {
	fmt.Printf("[TEST]"+format+"\n", v...)
}
