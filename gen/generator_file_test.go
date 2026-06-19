package gen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aviate-labs/agent-go/gen"
)

func TestNewGeneratorFromFile_imports(t *testing.T) {
	root := filepath.Join(t.TempDir(), "proj")
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("shared/ids.did", `type Id = text;`)
	write("candid/user.did",
		"import \"../shared/ids.did\";\n"+
			"type User = record { id : Id };\n"+
			"service : { get_user : (Id) -> (User) query };")
	write("backend.did",
		"import service \"./candid/user.did\";\n"+
			"service : { ping : () -> () };")

	g, err := gen.NewGeneratorFromFile("backend", "backend", "backend", filepath.Join(root, "backend.did"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	out := string(raw)
	for _, want := range []string{"type User struct", "func (a BackendAgent) GetUser", "func (a BackendAgent) Ping"} {
		if !strings.Contains(out, want) {
			t.Errorf("generated output missing %q\n---\n%s", want, out)
		}
	}
}
