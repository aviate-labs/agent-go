package did

import "testing"

func TestParseDID_quotedFieldLabel(t *testing.T) {
	d, err := ParseDID([]rune(`type T = record { "principal" : principal };`))
	if err != nil {
		t.Fatal(err)
	}
	td, ok := d.Definitions[0].(Type)
	if !ok {
		t.Fatalf("definition 0 is %T, want Type", d.Definitions[0])
	}
	rec, ok := td.Data.(Record)
	if !ok {
		t.Fatalf("data is %T, want Record", td.Data)
	}
	if rec[0].Name == nil {
		t.Fatal("field name is nil")
	}
	if got := *rec[0].Name; got != "principal" {
		t.Fatalf("field name = %q, want %q (surrounding quotes not stripped)", got, "principal")
	}
}

func TestParseDID_quotedMethodName(t *testing.T) {
	d, err := ParseDID([]rune(`service : { "my-method" : () -> () }`))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Services) == 0 {
		t.Fatal("no services parsed")
	}
	if got := d.Services[0].Methods[0].Name; got != "my-method" {
		t.Fatalf("method name = %q, want %q (surrounding quotes not stripped)", got, "my-method")
	}
}

func TestParseDID_quotedArgName(t *testing.T) {
	d, err := ParseDID([]rune(`type F = func ("arg" : nat) -> ();`))
	if err != nil {
		t.Fatal(err)
	}
	td := d.Definitions[0].(Type)
	fn := td.Data.(Func)
	if got := *fn.ArgTypes[0].Name; got != "arg" {
		t.Fatalf("arg name = %q, want %q (surrounding quotes not stripped)", got, "arg")
	}
}
