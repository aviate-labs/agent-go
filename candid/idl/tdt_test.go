package idl_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/did"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/internal/ctest"
)

// candidTestFile resolves a conformance vector from the upstream checkout the
// flake pins; a vendored copy would drift from the pin undetected.
func candidTestFile(t *testing.T, name string) string {
	t.Helper()
	dir := os.Getenv("CANDID_TEST_DIR")
	if dir == "" {
		t.Skip("CANDID_TEST_DIR unset: run via `nix develop` for the conformance vectors")
	}
	return filepath.Join(dir, name)
}

// Every vector file upstream ships, so a new one starts running when the pin
// moves instead of being silently missed.
func TestConformanceVectors(t *testing.T) {
	files, err := filepath.Glob(candidTestFile(t, "*.test.did"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no conformance vectors found")
	}
	for _, f := range files {
		t.Run(strings.TrimSuffix(filepath.Base(f), ".test.did"), func(t *testing.T) {
			runVectors(t, f)
		})
	}
}

func runVectors(t *testing.T, path string) {
	rawDid, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ctest.NewTestParser([]rune(string(rawDid)))
	if err != nil {
		t.Fatal(err)
	}
	n, err := p.ParseEOF(ctest.TestData)
	if err != nil {
		t.Fatal(err)
	}
	defs := make(map[string]did.Data)
	for _, n := range n.Children() {
		if n.Name != ctest.Def.Name {
			continue
		}
		cs := n.Children()
		defs[cs[0].Value()] = did.ConvertDataNode(cs[1].Children()[0])
	}
	env := did.NewTypeEnv(defs)

	for _, n := range n.Children() {
		switch n.Name {
		case ctest.CommentText.Name, ctest.Def.Name: // ignore
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
					ctest.TestTest.Name,
					ctest.TestNot.Name:
					test = n
				case ctest.Description.Name:
					desc = strings.Trim(n.Value(), `"`)
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
			ts, vs, err := candid.Decode(in)
			if err == nil {
				err = checkAgainst(env, test, ts, vs)
			}
			switch test.Name {
			case ctest.TestBad.Name:
				if err == nil {
					t.Errorf("%s: expected decode failure", desc)
				}
			case ctest.TestGood.Name, ctest.TestTest.Name, ctest.TestNot.Name:
				if err != nil {
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

// checkAgainst verifies the decoded types against the assertion's `: (type)`
// annotation.
func checkAgainst(env *did.TypeEnv, test *parser.Node, ts []idl.Type, vs []any) error {
	var anns []*parser.Node
	var walk func(n *parser.Node)
	walk = func(n *parser.Node) {
		if n.Name == ctest.AnnType.Name {
			anns = append(anns, n)
			return
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(test)

	// Arguments beyond the expected arity are ignored.
	for i, a := range anns {
		want, err := annotationType(env, a)
		if err != nil {
			return err
		}
		if i >= len(ts) {
			// A missing trailing argument reads as null where that is allowed.
			if err := idl.Subtype(new(idl.NullType), want); err != nil {
				return fmt.Errorf("argument %d: %w", i, err)
			}
			continue
		}
		if err := idl.Subtype(narrow(ts[i], vs[i]), want); err != nil {
			return fmt.Errorf("argument %d: %w", i, err)
		}
		if err := checkSelectedTag(vs[i], want); err != nil {
			return fmt.Errorf("argument %d: %w", i, err)
		}
	}
	return nil
}

func annotationType(env *did.TypeEnv, a *parser.Node) (idl.Type, error) {
	return env.IDLType(did.ConvertDataNode(a.Children()[0]))
}

// checkSelectedTag verifies the tag actually carried is one the expected type
// knows: Subtype only compares tags common to both.
func checkSelectedTag(v any, want idl.Type) error {
	variant, ok := v.(*idl.Variant)
	if !ok {
		return nil
	}
	u, ok := want.(*idl.VariantType)
	if !ok {
		return nil
	}
	for _, f := range u.Fields {
		if idl.LabelHash(f.Name) == idl.LabelHash(variant.Name) {
			return nil
		}
	}
	return fmt.Errorf("variant tag %s: not in the expected type", variant.Name)
}

// narrow reduces a wire type to the part the value occupies, so the comparison
// does not judge bytes that were never read.
func narrow(t idl.Type, v any) idl.Type {
	switch t := t.(type) {
	case *idl.VectorType:
		if vs, ok := v.([]any); ok && len(vs) == 0 {
			return idl.NewVectorType(new(idl.EmptyType))
		}
	case *idl.VariantType:
		sel, ok := v.(*idl.Variant)
		if !ok {
			return t
		}
		for _, f := range t.Fields {
			if idl.LabelHash(f.Name) != idl.LabelHash(sel.Name) {
				continue
			}
			return idl.NewVariantType(map[string]idl.Type{f.Name: narrow(f.Type, sel.Value)})
		}
	}
	return t
}
