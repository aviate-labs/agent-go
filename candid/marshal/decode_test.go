package marshal_test

import (
	"encoding/hex"
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
)

func ExampleUnmarshal_null() {
	var null idl.Null
	data, _ := hex.DecodeString("4449444c00017f")
	fmt.Println(marshal.Unmarshal(data, []any{&null}), null)
	// Output:
	// <nil> {}
}

func ExampleUnmarshal_optNat() {
	{ // 1
		var optNat *idl.Nat
		data, _ := hex.DecodeString("4449444c016e7d01000101")
		fmt.Println(marshal.Unmarshal(data, []any{&optNat}), optNat)
	}
	{ // null
		var optNat *idl.Nat
		data, _ := hex.DecodeString("4449444c016e7f010000")
		fmt.Println(marshal.Unmarshal(data, []any{&optNat}), optNat)
	}
	// Output:
	// <nil> 1
	// <nil> <nil>
}

func ExampleUnmarshal_principal() {
	var p principal.Principal
	data, _ := hex.DecodeString("4449444c0001680100")
	fmt.Println(marshal.Unmarshal(data, []any{&p}), p)
	// Output:
	// <nil> aaaaa-aa
}

func ExampleUnmarshal_record() {
	record := make(map[string]any)
	data, _ := hex.DecodeString("4449444c016c02d3e3aa027c868eb7027101002a04f09f92a9")
	fmt.Println(marshal.Unmarshal(data, []any{&record}), record)
	// Output:
	// <nil> map[4895187:42 5097222:💩]
}

func ExampleUnmarshal_struct() {
	var s struct {
		Foo string
		Bar idl.Int
	}
	data, _ := hex.DecodeString("4449444c016c02d3e3aa027c868eb7027101002a04f09f92a9")
	fmt.Println(marshal.Unmarshal(data, []any{&s}), s)
	// Output:
	// <nil> {💩 42}
}

func ExampleUnmarshal_variant() {
	var v *idl.Variant
	data, _ := hex.DecodeString("4449444c016b019cc2017d01000000")
	fmt.Println(marshal.Unmarshal(data, []any{&v}), v)
	// Output:
	// <nil> &{0 0 variant {24860:nat}}
}

func ExampleUnmarshal_vector() {
	var vec []any
	data, _ := hex.DecodeString("4449444c016d7c01000400010203")
	fmt.Println(marshal.Unmarshal(data, []any{&vec}), vec)
	// Output:
	// <nil> [0 1 2 3]
}

func TestUnmarshal_nat(t *testing.T) {
	data, err := marshal.Marshal([]any{idl.NewNat[uint](5)})
	if err != nil {
		t.Fatal(err)
	}
	var num idl.Nat
	if err := marshal.Unmarshal(data, []any{&num}); err != nil {
		t.Fatal(err)
	}
	if num.BigInt().Uint64() != 5 {
		t.Errorf("unexpected num: %s", num)
	}

	{ // uint8
		data, err := marshal.Marshal([]any{uint8(5)})
		if err != nil {
			t.Fatal(err)
		}
		var num uint8
		if err := marshal.Unmarshal(data, []any{&num}); err != nil {
			t.Fatal(err)
		}
		if num != 5 {
			t.Errorf("unexpected num: %d", num)
		}
	}
}

func TestUnmarshal_opt(t *testing.T) {
	var optNat *idl.Nat
	data, _ := hex.DecodeString("4449444c016e7d010000")
	if err := marshal.Unmarshal(data, []any{&optNat}); err != nil {
		t.Fatal(err)
	}
	if optNat != nil {
		t.Error(optNat)
	}
}

func TestUnmarshal_string_invalid(t *testing.T) {
	data, err := marshal.Marshal([]any{true})
	if err != nil {
		t.Fatal(err)
	}
	var name string
	if err := marshal.Unmarshal(data, []any{&name}); err == nil {
		t.Fatal(err)
	}
}

func TestUnmarshal_string_valid(t *testing.T) {
	data, err := marshal.Marshal([]any{"John"})
	if err != nil {
		t.Fatal(err)
	}
	var name string
	if err := marshal.Unmarshal(data, []any{name}); err == nil {
		t.Fatal()
	}
	if err := marshal.Unmarshal(data, []any{&name}); err != nil {
		t.Fatal(err)
	}
	if name != "John" {
		t.Errorf("unexpected name: %q", name)
	}

	{ // Multiple strings.
		data, err := marshal.Marshal([]any{"John", "Doe"})
		if err != nil {
			t.Fatal(err)
		}
		var firstName string
		var lastName string
		if err := marshal.Unmarshal(data, []any{&firstName, &lastName}); err != nil {
			t.Fatal(err)
		}
		if firstName != "John" {
			t.Errorf("unexpected first name: %q", firstName)
		}
		if lastName != "Doe" {
			t.Errorf("unexpected last name: %q", lastName)
		}
	}
}
