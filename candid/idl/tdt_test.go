package idl_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

// candidTestFile resolves a conformance vector from the upstream checkout the
// flake pins. Set by the nix dev shell, which CI runs under; a vendored copy is
// deliberately avoided since it drifts from the pin undetected.
func candidTestFile(t *testing.T, name string) string {
	t.Helper()
	dir := os.Getenv("CANDID_TEST_DIR")
	if dir == "" {
		t.Fatal("CANDID_TEST_DIR unset: run tests via `nix develop`")
	}
	return filepath.Join(dir, name)
}

// An empty argument list decodes fine on its own; whether it is valid depends
// on the expected type (`(nat)` must fail, `(opt nat)` must yield null), which
// an untyped decode cannot see. Unmarshal, which does take a type, currently
// rejects both: the opt case is a real conformance gap.
var untypedDecodeCannotJudge = map[string]bool{
	`"missing argument: nat   fails"`: true,
	`"missing argument: empty fails"`: true,
}

func TestTypeDefinitionTable(t *testing.T) {
	rawDid, err := os.ReadFile(candidTestFile(t, "prim.test.did"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := ctest.NewParser([]rune(string(rawDid)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := p.ParseEOF(ctest.TestData)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range n.Children() {
		switch n.Name {
		case ctest.CommentText.Name: // ignore
		case ctest.Test.Name:
			var (
				in   []byte
				test *parser.Node
				desc string
			)
			for _, n := range n.Children() {
				switch n.Name {
				case ctest.BlobInput.Name:
					b, err := parseBlob(n)
					if err != nil {
						t.Fatal(err)
					}
					in = b
				case ctest.TestBad.Name,
					ctest.TestGood.Name,
					ctest.TestTest.Name:
					test = n
				case ctest.Description.Name:
					desc = n.Value()
				case ctest.TextInput.Name: // textual value, no blob to decode
				default:
					t.Fatal(n)
				}
			}
			// Textual inputs carry no blob to decode. Per the upstream README an
			// implementation without a textual value parser skips those, and
			// treats a mixed `blob == text` comparison as a plain decode check.
			if in == nil {
				continue
			}
			// candid.Decode takes no expected type, so a vector that is only
			// invalid relative to its `: (type)` annotation cannot be judged
			// here. Reading those annotations needs the candid type grammar in
			// ctest, which it does not have yet.
			if untypedDecodeCannotJudge[desc] {
				continue
			}
			switch test.Name {
			case ctest.TestBad.Name:
				ts, as, err := candid.Decode(in)
				if err == nil {
					t.Errorf("%s: %v, %v", desc, ts, as)
				}
			case ctest.TestGood.Name, ctest.TestTest.Name:
				if _, _, err := candid.Decode(in); err != nil {
					t.Errorf("%s: %v", desc, err)
				}
			}
		default:
			t.Fatal(n)
		}
	}
}

func parseBlob(n *parser.Node) ([]byte, error) {
	if n.Name != ctest.BlobInput.Name {
		return nil, fmt.Errorf("invalid type: %s", n.Name)
	}

	var bs []byte
	for _, n := range n.Children() {
		switch n.Name {
		case ctest.BlobAlpha.Name:
			bs = append(bs, []byte(n.Value())...)
		case ctest.BlobHex.Name:
			h, err := hex.DecodeString(n.Value())
			if err != nil {
				return nil, err
			}
			bs = append(bs, h...)
		default:
			return nil, fmt.Errorf("invalid type: %s", n.Name)
		}
	}
	return bs, nil
}
