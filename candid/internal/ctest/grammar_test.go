package ctest_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

func TestData(t *testing.T) {
	rawDid, err := os.ReadFile("../../idl/testdata/prim.test.did")
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
