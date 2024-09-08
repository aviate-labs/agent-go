package agent_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/ic"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
)

var _ = new(testLogger)

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

	a, _ := agent.New(agent.DefaultConfig)
	if err := a.Query(ic.LEDGER_PRINCIPAL, "account_balance_dfx", []any{struct {
		Account string `json:"account"`
	}{
		Account: "9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d",
	}}, []any{&balance}); err != nil {
		fmt.Println(err)
	}
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
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	n, err := a.ReadStateCertificate(ic.REGISTRY_PRINCIPAL, [][]hashtree.Label{{hashtree.Label("subnet")}})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range hashtree.ListPaths(n, nil) {
		if len(path) == 3 && string(path[0]) == "subnet" && string(path[2]) == "public_key" {
			subnetID := principal.Principal{Raw: []byte(path[1])}
			_ = subnetID
		}
	}
}

func TestAgent_Call_provisionalTopUpCanister(t *testing.T) {
	a, err := ic0.NewAgent(ic.MANAGEMENT_CANISTER_PRINCIPAL, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.ProvisionalTopUpCanister(ic0.ProvisionalTopUpCanisterArgs{
		CanisterId: ic.LEDGER_PRINCIPAL,
	}); err == nil {
		t.Fatal()
	}
}

func TestAgent_Query_Ed25519(t *testing.T) {
	id, err := identity.NewRandomEd25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	a, _ := agent.New(agent.Config{
		Identity: id,
	})
	type Account struct {
		Account string `ic:"account"`
	}
	var balance struct {
		E8S uint64 `ic:"e8s"`
	}
	if err := a.Query(ic.LEDGER_PRINCIPAL, "account_balance_dfx", []any{
		Account{"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d"},
	}, []any{&balance}); err != nil {
		t.Fatal(err)
	}
}

func TestAgent_Query_Secp256k1(t *testing.T) {
	id, err := identity.NewRandomSecp256k1Identity()
	if err != nil {
		t.Fatal(err)
	}
	a, _ := agent.New(agent.Config{
		Identity: id,
	})
	type Account struct {
		Account string `ic:"account"`
	}
	var balance struct {
		E8S uint64 `ic:"e8s"`
	}
	if err := a.Query(ic.LEDGER_PRINCIPAL, "account_balance_dfx", []any{
		Account{"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d"},
	}, []any{&balance}); err != nil {
		t.Fatal(err)
	}
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

func (t testLogger) Printf(format string, v ...any) {
	fmt.Printf("[TEST]"+format+"\n", v...)
}

// Refer ic/wallet/README.md for more information
func Test_Agent_LocalNet(t *testing.T) {
	host, err := url.Parse("http://localhost:4943")
	if err != nil {
		panic(err)
	}
	cfg := agent.Config{
		ClientConfig:                   &agent.ClientConfig{Host: host},
		FetchRootKey:                   true,
		DisableSignedQueryVerification: true, //MUST BE TRUE TO ACCESS LOCAL REPLICA
	}
	a, err := agent.New(cfg)
	if err != nil {
		panic(err)
	}

	principal := principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai")

	var s1 string
	err = a.Query(principal, "greet", []any{}, []any{&s1})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s1:%v\n", s1)

	var s2 string
	err = a.Query(principal, "concat", []any{"hello", "world"}, []any{&s2})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s2:%v\n", s2)

	var s3 string
	err = a.Call(principal, "sha256", []any{"hello, world", uint32(2)}, []any{&s3}) //2's type should match with taht defined in hasher canister.
	if err != nil {
		panic(err)
	}
	fmt.Printf("s3:%v\n", s3)
}
