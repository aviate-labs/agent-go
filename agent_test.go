package agent_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
)

var (
	LEDGER_PRINCIPAL   = principal.MustDecode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	REGISTRY_PRINCIPAL = principal.MustDecode("rwlgt-iiaaa-aaaaa-aaaaa-cai")
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
	_ = a.Query(LEDGER_PRINCIPAL, "account_balance_dfx", []any{
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
	if err := a.Query(LEDGER_PRINCIPAL, "account_balance_dfx", []any{struct {
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
	n, err := a.ReadStateCertificate(REGISTRY_PRINCIPAL, [][]hashtree.Label{{hashtree.Label("subnet")}})
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

func TestAgent_Query_Callback(t *testing.T) {
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	ledgerCanisterID := principal.MustDecode("ryjl3-tyaaa-aaaaa-aaaba-cai")

	type GetBlocksArgs struct {
		Start  uint64 `ic:"start" json:"start"`
		Length uint64 `ic:"length" json:"length"`
	}

	type QueryArchiveFn struct {
		/* TODO! */
	}

	type ArchivedBlocksRange struct {
		Start    uint64         `ic:"start" json:"start"`
		Length   uint64         `ic:"length" json:"length"`
		Callback QueryArchiveFn `ic:"callback" json:"callback"`
	}

	type QueryBlocksResponse struct {
		ChainLength     uint64                `ic:"chain_length" json:"chain_length"`
		Certificate     *[]byte               `ic:"certificate,omitempty" json:"certificate,omitempty"`
		Blocks          []any                 `ic:"blocks" json:"blocks"`
		FirstBlockIndex uint64                `ic:"first_block_index" json:"first_block_index"`
		ArchivedBlocks  []ArchivedBlocksRange `ic:"archived_blocks" json:"archived_blocks"`
	}

	req, err := a.CreateCandidAPIRequest(
		agent.RequestTypeQuery,
		ledgerCanisterID,
		"query_blocks",
		GetBlocksArgs{
			Start:  123,
			Length: 1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	var out QueryBlocksResponse
	if err := req.Query([]any{&out}, false); err != nil {
		t.Error(err)
	}
	archive := out.ArchivedBlocks[0]
	if archive.Start != 123 || archive.Length != 1 {
		t.Error(archive)
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
	if err := a.Query(LEDGER_PRINCIPAL, "account_balance_dfx", []any{
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
	if err := a.Query(LEDGER_PRINCIPAL, "account_balance_dfx", []any{
		Account{"9523dc824aa062dcd9c91b98f4594ff9c6af661ac96747daef2090b7fe87037d"},
	}, []any{&balance}); err != nil {
		t.Fatal(err)
	}
}

type testLogger struct{}

func (t testLogger) Printf(format string, v ...any) {
	fmt.Printf("[TEST]"+format+"\n", v...)
}
