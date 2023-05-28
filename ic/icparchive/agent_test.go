package icparchive_test

import (
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/icparchive"
	"github.com/aviate-labs/agent-go/ic/icpledger"
	"testing"
	"time"
)

func ExampleAgent_GetBlocks() {
	ledger, _ := icpledger.NewAgent(ic.LEDGER_PRINCIPAL, agent.Config{})
	archives, _ := ledger.Archives()
	for _, archive := range archives.Archives {
		archive, _ := icparchive.NewAgent(archive.CanisterId, agent.Config{})
		blocks, _ := archive.GetBlocks(icparchive.GetBlocksArgs{Start: 0, Length: 1})
		block := blocks.Ok.Blocks[0]
		unix := time.Unix(int64(block.Timestamp.TimestampNanos/1_000_000_000), 0)
		op := block.Transaction.Operation.Mint
		to := hex.EncodeToString(op.To)
		amount := op.Amount.E8s / 10_000_000
		fmt.Printf("%s %s %d ICP\n", unix.UTC().String(), to, amount)
	}
	// Output:
	// 2021-05-06 19:17:10 +0000 UTC 529ea51c22e8d66e8302eabd9297b100fdb369109822248bb86939a671fbc55b 15431 ICP
}

func TestUnmarshal_operation(t *testing.T) {
	t.Run("Mint", func(t *testing.T) {
		var o icparchive.Operation
		data, _ := hex.DecodeString("4449444c046b01c2f5d59903016c02fbca0102d8a38ca80d036d7b6c01e0a9b3027801000004746573741027000000000000")
		if err := idl.Unmarshal(data, []any{&o}); err != nil {
			t.Fatal(err)
		}
		if o.Mint == nil {
			t.Fatal(o.Mint)
		}
		if o.Burn != nil && o.Transfer != nil && o.Approve == nil && o.TransferFrom != nil {
			t.Fatal(o)
		}
		op := o.Mint
		if string(op.To) != "test" {
			t.Error(op.To)
		}
		if op.Amount.E8s != 10_000 {
			t.Error(op.Amount)
		}
	})
	t.Run("Burn", func(t *testing.T) {
		var o icparchive.Operation
		data, _ := hex.DecodeString("4449444c046b01ef80e5df02016c02eaca8a9e0402d8a38ca80d036d7b6c01e0a9b3027801000004746573741027000000000000")
		if err := idl.Unmarshal(data, []any{&o}); err != nil {
			t.Fatal(err)
		}
		if o.Burn == nil {
			t.Fatal(o.Burn)
		}
		if o.Mint != nil && o.Transfer != nil && o.Approve == nil && o.TransferFrom != nil {
			t.Fatal(o)
		}
		op := o.Burn
		if string(op.From) != "test" {
			t.Error(op.From)
		}
		if op.Amount.E8s != 10_000 {
			t.Error(op.Amount)
		}
	})
	t.Run("Transfer", func(t *testing.T) {
		var o icparchive.Operation
		data, _ := hex.DecodeString("4449444c046b01cbd6fda00b016c04fbca0102c6fcb60203eaca8a9e0402d8a38ca80d036d7b6c01e0a9b302780100000474657374000000000000000004746573741027000000000000")
		if err := idl.Unmarshal(data, []any{&o}); err != nil {
			t.Fatal(err)
		}
		if o.Transfer == nil {
			t.Fatal(o.Transfer)
		}
		if o.Transfer != nil && o.Burn != nil && o.Approve == nil && o.TransferFrom != nil {
			t.Fatal(o)
		}
		op := o.Transfer
		if string(op.From) != "test" {
			t.Error(op.From)
		}
		if string(op.To) != "test" {
			t.Error(op.To)
		}
		if op.Amount.E8s != 10_000 {
			t.Error(op.Amount)
		}
		if op.Fee.E8s != 0 {
			t.Error(op.Fee)
		}
	})
	t.Run("Approve", func(t *testing.T) {
		var o icparchive.Operation
		data, _ := hex.DecodeString("4449444c046b01adfaedfb01016c05c6fcb60202eaca8a9e0403b98792ea077cdea7f7da0d7fcb96dcb40e036c01e0a9b302786d7b0100000000000000000000047465737490ce000474657374")
		if err := idl.Unmarshal(data, []any{&o}); err != nil {
			t.Fatal(err)
		}
		if o.Approve == nil {
			t.Fatal(o.Transfer)
		}
		if o.Mint != nil && o.Burn != nil && o.Transfer == nil && o.TransferFrom != nil {
			t.Fatal(o)
		}
		op := o.Approve
		if string(op.From) != "test" {
			t.Error(op.From)
		}
		if string(op.Spender) != "test" {
			t.Error(op.Spender)
		}
		if op.AllowanceE8s.BigInt().Int64() != 10_000 {
			t.Error(op.AllowanceE8s)
		}
		if op.Fee.E8s != 0 {
			t.Error(op.Fee)
		}
		if op.ExpiresAt != nil {
			t.Error(op.ExpiresAt)
		}
	})
}
