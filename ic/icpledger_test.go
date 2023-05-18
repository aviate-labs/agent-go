package ic_test

import (
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"testing"
	"time"
)

var (
	canisterId principal.Principal
	hostRaw    = "http://localhost:8000"
	host, _    = url.Parse(hostRaw)
)

func Example_accountBalance() {
	host, _ := url.Parse("https://icp0.io")
	a, _ := icpledger.NewAgent(ic.LEDGER_PRINCIPAL, agent.Config{
		ClientConfig: &agent.ClientConfig{Host: host},
		FetchRootKey: true,
	})
	name, _ := a.Name()
	fmt.Println(name.Name)
	// Output:
	// Internet Computer
}

func TestAgent(t *testing.T) {
	if os.Getenv("DFX") != "true" {
		t.SkipNow()
	}

	// Default account of the anonymous principal.
	defaultAccount := principal.AnonymousID.AccountIdentifier(principal.DefaultSubAccount)

	t.Run("account_balance ed25519", func(t *testing.T) {
		id, _ := identity.NewRandomEd25519Identity()
		a, _ := icpledger.NewAgent(canisterId, agent.Config{
			Identity: id,
			ClientConfig: &agent.ClientConfig{
				Host: host,
			},
			FetchRootKey: true,
		})
		tokens, err := a.AccountBalance(icpledger.AccountBalanceArgs{
			Account: defaultAccount[:],
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8s != 1 {
			t.Error(tokens)
		}
	})

	t.Run("account_balance secp256k1", func(t *testing.T) {
		id, _ := identity.NewRandomSecp256k1Identity()
		a, _ := icpledger.NewAgent(canisterId, agent.Config{
			Identity: id,
			ClientConfig: &agent.ClientConfig{
				Host: host,
			},
			FetchRootKey: true,
		})
		tokens, err := a.AccountBalance(icpledger.AccountBalanceArgs{
			Account: defaultAccount[:],
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8s != 1 {
			t.Error(tokens)
		}
	})

	a, _ := icpledger.NewAgent(canisterId, agent.Config{
		ClientConfig: &agent.ClientConfig{
			Host: host,
		},
		FetchRootKey: true,
	})
	t.Run("account_balance", func(t *testing.T) {
		tokens, err := a.AccountBalance(icpledger.AccountBalanceArgs{
			Account: defaultAccount[:],
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8s != 1 {
			t.Error(tokens)
		}
	})

	t.Run("transfer", func(t *testing.T) {
		p, _ := principal.Decode("aaaaa-aa")
		subAccount := principal.DefaultSubAccount[:]
		to := p.AccountIdentifier(principal.DefaultSubAccount)
		result, err := a.Transfer(icpledger.TransferArgs{
			Memo: 0,
			Amount: icpledger.Tokens{
				E8s: 100_000,
			},
			Fee: icpledger.Tokens{
				E8s: 10_000,
			},
			FromSubaccount: &subAccount,
			To:             to[:],
			CreatedAtTime: &icpledger.TimeStamp{
				TimestampNanos: uint64(time.Now().UnixNano()),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if *result.Ok != 1 {
			t.Error(result)
		}
	})
}

func init() {
	canisterIdsRaw, _ := os.ReadFile("testdata/.dfx/local/canister_ids.json")
	type CanisterIds struct {
		Example struct {
			IC string `json:"local"`
		} `json:"example"`
	}
	var canisterIds CanisterIds
	_ = json.Unmarshal(canisterIdsRaw, &canisterIds)
	canisterId, _ = principal.Decode(canisterIds.Example.IC)
}
