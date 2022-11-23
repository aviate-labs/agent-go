package idl_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/internal/blob"
	"github.com/aviate-labs/agent-go/candid/internal/candidtest"
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
)

func TestTypeDefinitionTable(t *testing.T) {
	rawDid, err := os.ReadFile("testdata/prim.test.did")
	if err != nil {
		t.Fatal(err)
	}
	p, err := ast.New(rawDid)
	if err != nil {
		t.Fatal(err)
	}
	n, err := candidtest.TestData(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		t.Error(err)
	}
	for _, n := range n.Children() {
		switch n.Type {
		case candidtest.CommentTextT: // ignore
		case candidtest.TestT:
			var (
				in   []byte
				test *ast.Node
				desc string
			)
			for _, n := range n.Children() {
				switch n.Type {
				case candidtest.BlobInputT:
					b, err := parseBlob(n)
					if err != nil {
						t.Fatal(err)
					}
					in = b
				case candidtest.TestBadT,
					candidtest.TestGoodT, // TODO
					candidtest.TestTestT: // TODO
					test = n
				case candidtest.DescriptionT:
					desc = n.Value
				default:
					t.Fatal(n)
				}
			}
			switch test.Type {
			case candidtest.TestBadT:
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

func parseBlob(n *ast.Node) ([]byte, error) {
	if n.Type != candidtest.BlobInputT {
		return nil, fmt.Errorf("invalid type: %s", n.TypeString())
	}
	if len(n.Value) == 0 {
		return []byte{}, nil
	}
	p, err := ast.New([]byte(n.Value))
	if err != nil {
		return nil, err
	}
	b, err := blob.Blob(p)
	if err != nil {
		return nil, err
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		return nil, err
	}
	var bs []byte
	for _, n := range b.Children() {
		switch n.Type {
		case blob.AlphaT:
			bs = append(bs, []byte(n.Value)...)
		case blob.HexT:
			h, err := hex.DecodeString(n.Value)
			if err != nil {
				return nil, err
			}
			bs = append(bs, h...)
		default:
			return nil, fmt.Errorf("invalid type: %s", n.TypeString())
		}
	}
	return bs, nil
}
