package marshal_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
)

func ExampleUnmarshal_optNat() {
	var optNat idl.Optional
	data, _ := hex.DecodeString("4449444c016e7d01000101")
	fmt.Println(marshal.Unmarshal(data, []any{&optNat}), optNat)
	// Output:
	// <nil> {1 nat}
}

func ExampleUnmarshal_record() {
	record := make(map[string]any)
	data, _ := hex.DecodeString("4449444c016c02d3e3aa027c868eb7027101002a04f09f92a9")
	fmt.Println(marshal.Unmarshal(data, []any{&record}), record)
	// Output:
	// <nil> map[4895187:42 5097222:💩]
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
	var optNat idl.Optional
	data, _ := hex.DecodeString("4449444c016e7d010000")
	if err := marshal.Unmarshal(data, []any{&optNat}); err != nil {
		t.Fatal(err)
	}
	if optNat.V != nil {
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
