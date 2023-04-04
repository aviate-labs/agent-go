package ledger_test

import (
	_ "embed"
	"encoding/json"
	"github.com/aviate-labs/agent-go/identity"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go/ledger"
	"github.com/aviate-labs/agent-go/principal"
)

var (
	canisterId principal.Principal
	hostRaw    = "http://localhost:8000"
	host, _    = url.Parse(hostRaw)
)

func TestAgent(t *testing.T) {
	if os.Getenv("DFX") != "true" {
		t.SkipNow()
	}

	t.Run("account_balance ed25519", func(t *testing.T) {
		id, _ := identity.NewRandomEd25519Identity()
		a := ledger.NewWithIdentity(canisterId, host, id)
		tokens, err := a.AccountBalance(ledger.AccountBalanceArgs{
			Account: principal.AnonymousID.AccountIdentifier(principal.DefaultSubAccount),
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8S != 1 {
			t.Error(tokens)
		}
	})

	t.Run("account_balance secp256k1", func(t *testing.T) {
		id, _ := identity.NewRandomSecp256k1Identity()
		a := ledger.NewWithIdentity(canisterId, host, id)
		tokens, err := a.AccountBalance(ledger.AccountBalanceArgs{
			Account: principal.AnonymousID.AccountIdentifier(principal.DefaultSubAccount),
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8S != 1 {
			t.Error(tokens)
		}
	})

	a := ledger.New(canisterId, host)
	t.Run("account_balance", func(t *testing.T) {
		tokens, err := a.AccountBalance(ledger.AccountBalanceArgs{
			Account: principal.AnonymousID.AccountIdentifier(principal.DefaultSubAccount),
		})
		if err != nil {
			t.Fatal(err)
		}
		if tokens.E8S != 1 {
			t.Error(tokens)
		}
	})

	t.Run("transfer", func(t *testing.T) {
		p, _ := principal.Decode("aaaaa-aa")
		subAccount := ledger.SubAccount(principal.DefaultSubAccount)
		tokens, err := a.Transfer(ledger.TransferArgs{
			Memo: 0,
			Amount: ledger.Tokens{
				E8S: 100_000,
			},
			Fee: ledger.Tokens{
				E8S: 10_000,
			},
			FromSubAccount: &subAccount,
			To:             p.AccountIdentifier(principal.DefaultSubAccount),
			CreatedAtTime: &ledger.TimeStamp{
				TimestampNanos: uint64(time.Now().UnixNano()),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if *tokens != 1 {
			t.Error(tokens)
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
