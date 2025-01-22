package idl_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/internal/blob"
	"github.com/aviate-labs/agent-go/candid/internal/candidtest"
)

func TestTypeDefinitionTable(t *testing.T) {
	rawDid, err := os.ReadFile("testdata/prim.test.did")
	if err != nil {
		t.Fatal(err)
	}
	p, err := candidtest.NewParser([]rune(string(rawDid)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := p.ParseEOF(candidtest.TestData)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range n.Children() {
		switch n.Name {
		case candidtest.CommentText.Name: // ignore
		case candidtest.Test.Name:
			var (
				in   []byte
				test *parser.Node
				desc string
			)
			for _, n := range n.Children() {
				switch n.Name {
				case candidtest.BlobInput.Name:
					b, err := parseBlob(n)
					if err != nil {
						t.Fatal(err)
					}
					in = b
				case candidtest.TestBad.Name,
					candidtest.TestGood.Name,
					candidtest.TestTest.Name:
					test = n
				case candidtest.Description.Name:
					desc = n.Value()
				default:
					t.Fatal(n)
				}
			}
			switch test.Name {
			case candidtest.TestBad.Name:
				ts, as, err := idl.Decode(in)
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
	if n.Name != candidtest.BlobInput.Name {
		return nil, fmt.Errorf("invalid type: %s", n.Name)
	}
	if len(n.Value()) == 0 {
		return []byte{}, nil
	}
	p, err := parser.New([]rune(n.Value()))
	if err != nil {
		return nil, err
	}
	b, err := p.ParseEOF(blob.Blob)
	if err != nil {
		return nil, err
	}

	var bs []byte
	for _, n := range b.Children() {
		switch n.Name {
		case blob.Alpha.Name:
			bs = append(bs, []byte(n.Value())...)
		case blob.Hex.Name:
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
