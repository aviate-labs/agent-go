package ledger_test

import (
	"os"
	"testing"
)

func checkEnabled(t *testing.T) {
	// The reason for this is that the tests are very slow.
	if os.Getenv("LEDGER_TEST_ENABLE") != "true" {
		t.Skip("Skipping registry tests. Set LEDGER_TEST_ENABLE=true to enable.")
	}
}
