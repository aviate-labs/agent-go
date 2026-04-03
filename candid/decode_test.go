package candid

import (
	"math/big"
	"testing"

	"github.com/niccolofant/agent-go/candid/idl"
)

func TestDecodeRaw(t *testing.T) {
	type Example struct {
		N idl.Nat `ic:"nat"`
	}
	bi := big.NewInt(100)
	e, err := Marshal([]any{Example{N: idl.NewBigNat(bi)}})
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := Unmarshal(e, []any{&v}); err != nil {
		t.Fatal(err)
	}
	type ExampleRaw struct {
		N idl.RawMessage `ic:"nat"`
	}
	var r ExampleRaw
	if err := Unmarshal(e, []any{&r}); err != nil {
		t.Fatal(err)
	}
	var n idl.Nat
	if err := r.N.Unmarshal(&n); err != nil {
		t.Fatal(err)
	}
	if n.BigInt() == nil || n.BigInt().Cmp(bi) != 0 {
		t.Error(n)
	}
}

func TestDecode_futureTypeOpcode(t *testing.T) {
	t.Run("unreferenced future type", func(t *testing.T) {
		wire := []byte{
			'D', 'I', 'D', 'L',
			0x01,       // type table count = 1
			0x67, 0x00, // future-type opcode -25, body length 0
			0x01, // arg count = 1
			0x7d, // arg type = nat (-3)
			0x2a, // nat 42
		}

		var n idl.Nat
		if err := Unmarshal(wire, []any{&n}); err != nil {
			t.Fatalf("decode failed on future type: %v", err)
		}
		if n.BigInt().Cmp(big.NewInt(42)) != 0 {
			t.Fatalf("got %v, want 42", n)
		}
	})

	t.Run("two args, second is a future value", func(t *testing.T) {
		// Forward-compat: a payload with two args, where the second arg has
		// a future type. The decoder must consume the future value bytes
		// (<m> <n> <m bytes>) so the byte stream is exhausted cleanly.
		wire := []byte{
			'D', 'I', 'D', 'L',
			0x01,       // type table count = 1
			0x67, 0x00, // future-type opcode -25, body length 0
			0x02,             // arg count = 2
			0x7d,             // arg 0 type = nat
			0x00,             // arg 1 type = type-table index 0 (future)
			0x2a,             // nat 42
			0x03,             // future value: m=3
			0x00,             // future value: n=0
			0xaa, 0xbb, 0xcc, // body
		}

		var n idl.Nat
		var f any
		if err := Unmarshal(wire, []any{&n, &f}); err != nil {
			t.Fatalf("decode failed on future value: %v", err)
		}
		if n.BigInt().Cmp(big.NewInt(42)) != 0 {
			t.Fatalf("got %v, want 42", n)
		}
	})
}

func TestDecode_recordFieldOrdering(t *testing.T) {
	t.Run("duplicate id rejected", func(t *testing.T) {
		wire := []byte{
			'D', 'I', 'D', 'L',
			0x01,       // type table count = 1
			0x6c,       // record opcode (-20)
			0x02,       // field count = 2
			0x05, 0x7d, // id 5, type nat
			0x05, 0x7d, // id 5 again - duplicate
			0x01,       // arg count = 1
			0x00,       // arg type = type-table index 0
			0x05, 0x2a, // record fields: id=5, nat=42
		}
		var m map[string]any
		if err := Unmarshal(wire, []any{&m}); err == nil {
			t.Fatal("expected error for duplicate field id, got nil")
		}
	})

	t.Run("out-of-order ids rejected", func(t *testing.T) {
		wire := []byte{
			'D', 'I', 'D', 'L',
			0x01,       // type table count = 1
			0x6c,       // record opcode
			0x02,       // field count = 2
			0x05, 0x7d, // id 5, type nat
			0x03, 0x7d, // id 3 (out of order: 3 < 5)
			0x01,
			0x00,
			0x2a, 0x2a,
		}
		var m map[string]any
		if err := Unmarshal(wire, []any{&m}); err == nil {
			t.Fatal("expected error for unordered field ids, got nil")
		}
	})
}
