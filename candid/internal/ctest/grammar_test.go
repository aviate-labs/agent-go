package ctest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

func TestData(t *testing.T) {
	dir := os.Getenv("CANDID_TEST_DIR")
	if dir == "" {
		t.Skip("CANDID_TEST_DIR unset: run via `nix develop` for the conformance vectors")
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.test.did"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no conformance vectors found")
	}
	for _, f := range files {
		t.Run(strings.TrimSuffix(filepath.Base(f), ".test.did"), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			p, err := ctest.NewTestParser(bytes.Runes(raw))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := p.ParseEOF(ctest.TestData); err != nil {
				t.Error(err)
			}
		})
	}
}
