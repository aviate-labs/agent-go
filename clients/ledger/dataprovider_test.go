package ledger_test

import (
	"github.com/aviate-labs/agent-go/clients/ledger"
	"testing"
)

func TestDataProvider_GetRawBlock(t *testing.T) {
	checkEnabled(t)

	dp, err := ledger.NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp.GetRawBlock(0); err != nil {
		t.Error(err)
	}
}

func TestDataProvider_GetRawBlocks(t *testing.T) {
	checkEnabled(t)

	dp, err := ledger.NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	n := 3 * ledger.MaxBlocksPerRequest
	blocks, err := dp.GetRawBlocks(0, ledger.BlockIndex(n))
	if err != nil {
		t.Error(err)
	}
	if len(blocks) != n {
		t.Errorf("expected %d blocks, got %d", n, len(blocks))
	}
}

func TestDataProvider_GetTipOfChain(t *testing.T) {
	checkEnabled(t)

	dp, err := ledger.NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dp.GetTipOfChain(); err != nil {
		t.Error(err)
	}
}
