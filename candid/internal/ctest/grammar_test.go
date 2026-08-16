package ctest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

func TestData(t *testing.T) {
	dir := os.Getenv("CANDID_TEST_DIR")
	if dir == "" {
		t.Fatal("CANDID_TEST_DIR unset: run tests via `nix develop`")
	}
	rawDid, err := os.ReadFile(filepath.Join(dir, "prim.test.did"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := ctest.NewParser(bytes.Runes(rawDid))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(ctest.TestData); err != nil {
		t.Fatal(err)
	}
}
