package did

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasType(d *Description, id string) bool {
	for _, def := range d.Definitions {
		if t, ok := def.(Type); ok && t.Id == id {
			return true
		}
	}
	return false
}

func methodNames(d *Description) []string {
	var names []string
	for _, s := range d.Services {
		for _, m := range s.Methods {
			names = append(names, m.Name)
		}
	}
	return names
}

func TestParseDIDFile_imports(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "proj")

	// Shared types live one level up from the service partials.
	writeFile(t, filepath.Join(root, "shared", "ids.did"), `type Id = text;`)
	writeFile(t, filepath.Join(root, "shared", "err.did"),
		"import \"./ids.did\";\ntype Err = record { id : Id; msg : text };")

	// A service partial that imports shared types via ./ and ../.
	writeFile(t, filepath.Join(root, "candid", "user.did"),
		"import \"../shared/ids.did\";\nimport \"../shared/err.did\";\n"+
			"type User = record { id : Id };\n"+
			"service : { get_user : (Id) -> (User) query };")

	// Aggregate: import service for methods, plain import for types only.
	writeFile(t, filepath.Join(root, "backend.did"),
		"import service \"./candid/user.did\";\n"+
			"import \"./shared/err.did\";\n"+
			"service : { ping : () -> () };")

	d, err := ParseDIDFile(filepath.Join(root, "backend.did"))
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"Id", "Err", "User"} {
		if !hasType(d, id) {
			t.Errorf("missing merged type %q", id)
		}
	}

	got := map[string]bool{}
	for _, n := range methodNames(d) {
		got[n] = true
	}
	for _, want := range []string{"ping", "get_user"} {
		if !got[want] {
			t.Errorf("missing merged method %q; got %v", want, methodNames(d))
		}
	}
}

// A file plain-imported via one path and `import service` via another must
// still contribute its methods: the service flag is the union over all paths.
func TestParseDIDFile_asymmetricServiceImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "d.did"), "service : { d_method : () -> () };")
	writeFile(t, filepath.Join(dir, "b.did"),
		"import \"./d.did\";\nservice : { b_method : () -> () };")
	writeFile(t, filepath.Join(dir, "c.did"),
		"import service \"./d.did\";\nservice : { c_method : () -> () };")
	writeFile(t, filepath.Join(dir, "a.did"),
		"import service \"./b.did\";\nimport service \"./c.did\";\n"+
			"service : { a_method : () -> () };")

	d, err := ParseDIDFile(filepath.Join(dir, "a.did"))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, n := range methodNames(d) {
		got[n] = true
	}
	for _, want := range []string{"a_method", "b_method", "c_method", "d_method"} {
		if !got[want] {
			t.Errorf("missing merged method %q; got %v", want, methodNames(d))
		}
	}
}

func TestParseDIDFile_conflictingType(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "x.did"), "type T = text;")
	writeFile(t, filepath.Join(dir, "y.did"), "type T = nat;")
	writeFile(t, filepath.Join(dir, "main.did"),
		"import \"./x.did\";\nimport \"./y.did\";\nservice : { f : () -> () };")

	if _, err := ParseDIDFile(filepath.Join(dir, "main.did")); err == nil {
		t.Fatal("expected error on conflicting type T definitions")
	}
}

// An identical re-declaration of a type through different import paths is not a
// conflict.
func TestParseDIDFile_duplicateTypeOK(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "x.did"), "type T = text;")
	writeFile(t, filepath.Join(dir, "y.did"), "import \"./x.did\";\ntype Y = T;")
	writeFile(t, filepath.Join(dir, "main.did"),
		"import \"./x.did\";\nimport \"./y.did\";\nservice : { f : () -> () };")

	d, err := ParseDIDFile(filepath.Join(dir, "main.did"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasType(d, "T") || !hasType(d, "Y") {
		t.Errorf("expected types T and Y; got %v", d.Definitions)
	}
}

// A cycle whose two ends reach the same file via different paths (one through a
// symlinked dir) must be broken by canonicalizing the key, not loop forever.
func TestParseDIDFile_symlinkCycle(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink(dir, filepath.Join(dir, "link")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	writeFile(t, filepath.Join(dir, "a.did"), "import \"./link/a.did\";\ntype A = text;")

	done := make(chan struct{})
	var d *Description
	var err error
	go func() {
		d, err = ParseDIDFile(filepath.Join(dir, "a.did"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ParseDIDFile did not terminate; symlink alias defeated the cycle guard")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !hasType(d, "A") {
		t.Error("missing type A")
	}
}

func TestParseDIDFile_cycle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.did"), "import \"./b.did\";\ntype A = text;")
	writeFile(t, filepath.Join(dir, "b.did"), "import \"./a.did\";\ntype B = text;")

	d, err := ParseDIDFile(filepath.Join(dir, "a.did"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasType(d, "A") || !hasType(d, "B") {
		t.Errorf("expected both A and B after cyclic import resolution")
	}
}
