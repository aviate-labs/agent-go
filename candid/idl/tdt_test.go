package idl_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

func TestTypeDefinitionTable(t *testing.T) {
	rawDid, err := os.ReadFile("testdata/prim.test.did")
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
				default:
					t.Fatal(n)
				}
			}
			switch test.Name {
			case ctest.TestBad.Name:
				ts, as, err := candid.Decode(in)
				if err == nil {
					t.Errorf("%s: %v, %v", desc, ts, as)
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
