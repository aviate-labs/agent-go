package did

import (
	"path/filepath"
	"testing"
)

// Candid permits // line comments and /* */ block comments anywhere whitespace
// is allowed. These pin that goic tolerates both, including a // comment that
// ends at EOF with no trailing newline, and comments inside imported partials.
func TestParseDID_comments(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"full line", "// header\ntype T = text;\n"},
		{"trailing", "type T = text; // alias\n"},
		{"block", "/* header */\ntype T = text;\n"},
		{"block multiline", "/* multi\n line */\ntype T = text;\n"},
		{"block star before close", "/* a **/\ntype T = text;\n"},
		{"block only stars", "/***/\ntype T = text;\n"},
		{"line comment at EOF no newline", "type T = text;\n// trailing"},
		{"only line comment at EOF", "// just a comment"},
		{"non ascii in comment", "// unsafe — documentation only\ntype T = text;\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ParseDID([]rune(c.src)); err != nil {
				t.Fatalf("ParseDID(%q): %v", c.src, err)
			}
		})
	}
}

// Comments surface in imported partials; the import path must tolerate them.
func TestParseDIDFile_commentsInImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "ids.did"),
		"// Type-only file.\n// Aliases are structural documentation only.\n"+
			"type UserId = text;\ntype OrgId = text; // trailing")
	writeFile(t, filepath.Join(dir, "backend.did"),
		"import \"./ids.did\";\n/* main service */\n"+
			"service : { whoami : () -> (UserId) query };")

	d, err := ParseDIDFile(filepath.Join(dir, "backend.did"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"UserId": false, "OrgId": false}
	for _, def := range d.Definitions {
		if ty, ok := def.(Type); ok {
			if _, tracked := want[ty.Id]; tracked {
				want[ty.Id] = true
			}
		}
	}
	for id, got := range want {
		if !got {
			t.Errorf("missing imported type %q", id)
		}
	}
	if got := methodNames(d); len(got) != 1 || got[0] != "whoami" {
		t.Errorf("methods = %v, want [whoami]", got)
	}
}
