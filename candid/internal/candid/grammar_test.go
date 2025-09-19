package candid_test

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/0x51-dev/upeg/parser"
	"github.com/0x51-dev/upeg/parser/op"
	"github.com/aviate-labs/agent-go/candid/internal/candid"
)

//go:embed testdata
var testdata embed.FS

func ExampleActorType() {
	var example = `{
	addUser : (name : text, age : nat8) -> (id : nat64);
	userName : (id : nat64) -> (text) query;
	userAge : (id : nat64) -> (nat8) query;
	deleteUser : (id : nat64) -> () oneway;
}`
	p, _ := candid.NewParser([]rune(example))
	actor, _ := p.ParseEOF(candid.ActorType)
	fmt.Println(len(actor.Children()))
	// output:
	// 4
}

func ExampleArgType() {
	p := func(s string) *parser.Parser {
		p, _ := candid.NewParser([]rune(s))
		return p
	}
	fmt.Println(p("name : text").ParseEOF(candid.ArgType))
	fmt.Println(p("age : nat8").ParseEOF(candid.ArgType))
	fmt.Println(p("id : nat64").ParseEOF(candid.ArgType))
	// output:
	// {"ArgType": [{"Id": "name"}, {"PrimType": "text"}]} <nil>
	// {"ArgType": [{"Id": "age"}, {"PrimType": "nat8"}]} <nil>
	// {"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]} <nil>
}

func ExampleComment() {
	var example = `// This is a comment.
`
	p, _ := candid.NewParser([]rune(example))
	comment, _ := p.ParseEOF(op.Capture{Value: candid.Comment})
	fmt.Println(comment.Children()[0].Value())
	// output:
	// This is a comment.
}

func ExampleConsType() {
	for _, record := range []string{
		"record {\n  num : nat;\n}",
		"record { nat; nat }",
		"record { 0 : nat; 1 : nat }",
	} {
		p, _ := candid.NewParser([]rune(record))
		fmt.Println(p.ParseEOF(candid.ConsType))
	}
	// output:
	// {"Record": [{"FieldType": [{"Id": "num"}, {"PrimType": "nat"}]}]} <nil>
	// {"Record": [{"FieldType": [{"PrimType": "nat"}]}, {"FieldType": [{"PrimType": "nat"}]}]} <nil>
	// {"Record": [{"FieldType": [{"Nat": "0"}, {"PrimType": "nat"}]}, {"FieldType": [{"Nat": "1"}, {"PrimType": "nat"}]}]} <nil>
}

func ExampleDef() {
	for _, def := range []string{
		"type list = opt node",
		"type color = variant { red; green; blue }",
		"type tree = variant {\n  leaf : int;\n  branch : record {left : tree; val : int; right : tree};\n}",
		"type stream = opt record {head : nat; next : func () -> stream}",
	} {
		p, _ := candid.NewParser([]rune(def))
		fmt.Println(p.ParseEOF(candid.Def))
	}
	// output:
	// {"Type": [{"Id": "list"}, {"Opt": [{"Id": "node"}]}]} <nil>
	// {"Type": [{"Id": "color"}, {"Variant": [{"FieldType": [{"Id": "red"}]}, {"FieldType": [{"Id": "green"}]}, {"FieldType": [{"Id": "blue"}]}]}]} <nil>
	// {"Type": [{"Id": "tree"}, {"Variant": [{"FieldType": [{"Id": "leaf"}, {"PrimType": "int"}]}, {"FieldType": [{"Id": "branch"}, {"Record": [{"FieldType": [{"Id": "left"}, {"Id": "tree"}]}, {"FieldType": [{"Id": "val"}, {"PrimType": "int"}]}, {"FieldType": [{"Id": "right"}, {"Id": "tree"}]}]}]}]}]} <nil>
	// {"Type": [{"Id": "stream"}, {"Opt": [{"Record": [{"FieldType": [{"Id": "head"}, {"PrimType": "nat"}]}, {"FieldType": [{"Id": "next"}, {"Func": [{"FuncType": [{"TupType": "()"}, {"TupType": [{"ArgType": [{"Id": "stream"}]}]}]}]}]}]}]}]} <nil>
}

func ExampleFuncType() {
	for _, function := range []string{
		"(text, text, nat16) -> (text, nat64)",
		"(name : text, address : text, nat16) -> (text, id : nat64)",
		"(name : text, address : text, nr : nat16) -> (nick : text, id : nat64)",
	} {
		p, _ := candid.NewParser([]rune(function))
		fmt.Println(p.ParseEOF(candid.FuncType))
	}
	// output:
	// {"FuncType": [{"TupType": [{"ArgType": [{"PrimType": "text"}]}, {"ArgType": [{"PrimType": "text"}]}, {"ArgType": [{"PrimType": "nat16"}]}]}, {"TupType": [{"ArgType": [{"PrimType": "text"}]}, {"ArgType": [{"PrimType": "nat64"}]}]}]} <nil>
	// {"FuncType": [{"TupType": [{"ArgType": [{"Id": "name"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "address"}, {"PrimType": "text"}]}, {"ArgType": [{"PrimType": "nat16"}]}]}, {"TupType": [{"ArgType": [{"PrimType": "text"}]}, {"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}]} <nil>
	// {"FuncType": [{"TupType": [{"ArgType": [{"Id": "name"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "address"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "nr"}, {"PrimType": "nat16"}]}]}, {"TupType": [{"ArgType": [{"Id": "nick"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}]} <nil>
}

func ExampleMethType() {
	for _, method := range []string{
		"addUser : (name : text, age : nat8) -> (id : nat64)",
		"userName : (id : nat64) -> (text) query",
		"userAge : (id : nat64) -> (nat8) query",
		"deleteUser : (id : nat64) -> () oneway",
	} {
		p, _ := candid.NewParser([]rune(method))
		fmt.Println(p.ParseEOF(candid.MethType))
	}
	// output:
	// {"MethType": [{"Id": "addUser"}, {"FuncType": [{"TupType": [{"ArgType": [{"Id": "name"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "age"}, {"PrimType": "nat8"}]}]}, {"TupType": [{"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}]}]} <nil>
	// {"MethType": [{"Id": "userName"}, {"FuncType": [{"TupType": [{"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}, {"TupType": [{"ArgType": [{"PrimType": "text"}]}]}, {"FuncAnn": "query"}]}]} <nil>
	// {"MethType": [{"Id": "userAge"}, {"FuncType": [{"TupType": [{"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}, {"TupType": [{"ArgType": [{"PrimType": "nat8"}]}]}, {"FuncAnn": "query"}]}]} <nil>
	// {"MethType": [{"Id": "deleteUser"}, {"FuncType": [{"TupType": [{"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]}, {"TupType": "()"}, {"FuncAnn": "oneway"}]}]} <nil>
}

func ExampleTupType() {
	for _, tuple := range []string{
		"(name : text, age : nat8)",
		"(id : nat64)",
		"()",
	} {
		p, _ := candid.NewParser([]rune(tuple))
		n, err := p.ParseEOF(candid.TupType)
		fmt.Println(n, err)
	}
	// output:
	// {"TupType": [{"ArgType": [{"Id": "name"}, {"PrimType": "text"}]}, {"ArgType": [{"Id": "age"}, {"PrimType": "nat8"}]}]} <nil>
	// {"TupType": [{"ArgType": [{"Id": "id"}, {"PrimType": "nat64"}]}]} <nil>
	// {"TupType": "()"} <nil>
}

func TestComment(t *testing.T) {
	var example = `// \ <-- oops
`
	p, err := candid.NewParser([]rune(example))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(op.Capture{Value: candid.Comment}); err != nil {
		t.Fatal(err)
	}
}

func TestDef_function(t *testing.T) {
	var example = `type engine = service {
  search : (query : text, callback : func (vec result) -> ());
}`
	p, err := candid.NewParser([]rune(example))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(candid.Def); err != nil {
		t.Fatal(err)
	}
}

func TestDef_service(t *testing.T) {
	var example = `type broker = service {
  findCounterService : (name : text) ->
    (service {up : () -> (); current : () -> nat});
}`
	p, err := candid.NewParser([]rune(example))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(candid.Def); err != nil {
		t.Fatal(err)
	}
}

func TestExamples(t *testing.T) {
	examples, _ := testdata.ReadDir("testdata")
	for _, example := range examples {
		t.Run(strings.TrimSuffix(example.Name(), ".did"), func(t *testing.T) {
			path := fmt.Sprintf("testdata/%s", example.Name())
			raw, _ := fs.ReadFile(testdata, path)
			p, err := candid.NewParser([]rune(string(raw)))
			if err != nil {
				t.Fatal(err)
			}
			n, err := p.ParseEOF(candid.Prog)
			if err != nil {
				t.Fatal(n, err)
			}
		})
	}
}

func TestName(t *testing.T) {
	for _, name := range []string{
		"addUser", "userName", "userAge", "deleteUser",
	} {
		p, err := candid.NewParser([]rune(name))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ParseEOF(candid.Name); err != nil {
			t.Error(err)
		}
	}
}

func TestWs(t *testing.T) {
	p, err := candid.NewParser([]rune("\n  \t\n "))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseEOF(candid.Ws); err != nil {
		t.Error(err)
	}
}
