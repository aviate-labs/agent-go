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

func TestUnmarshalDirect_skipsUnknownNestedFields(t *testing.T) {
	type ignoredVariant struct {
		Off *idl.Null `ic:"off,variant"`
		On  *uint64   `ic:"on,variant"`
	}
	type ignoredInner struct {
		A      uint64         `ic:"a"`
		Choice ignoredVariant `ic:"choice"`
		Maybe  *uint32        `ic:"maybe"`
		Names  []string       `ic:"names"`
	}
	type wire struct {
		Ignored []ignoredInner `ic:"a"`
		Tail    string         `ic:"z"`
	}
	type got struct {
		Tail string `ic:"z"`
	}

	maybe := uint32(7)
	on := uint64(99)
	encoded, err := Marshal([]any{wire{
		Ignored: []ignoredInner{
			{
				A:      1,
				Choice: ignoredVariant{On: &on},
				Maybe:  &maybe,
				Names:  []string{"alpha", "beta", "gamma"},
			},
			{
				A:      2,
				Choice: ignoredVariant{Off: new(idl.Null)},
				Names:  []string{"delta"},
			},
		},
		Tail: "kept",
	}})
	if err != nil {
		t.Fatal(err)
	}

	ts, _, err := decodeTypes(encoded)
	if err != nil {
		t.Fatal(err)
	}
	var value got
	if !canUnmarshalDirect(ts[0], &value) {
		t.Fatal("expected direct unmarshal path")
	}
	if err := Unmarshal(encoded, []any{&value}); err != nil {
		t.Fatal(err)
	}
	if value.Tail != "kept" {
		t.Fatalf("got Tail=%q, want kept", value.Tail)
	}
}

func TestUnmarshalDirect_unknownVariantArmStillErrors(t *testing.T) {
	type wireVariant struct {
		Known *uint64 `ic:"known,variant"`
		Other *uint64 `ic:"other,variant"`
	}
	type gotVariant struct {
		Known *uint64 `ic:"known,variant"`
	}

	other := uint64(42)
	encoded, err := Marshal([]any{wireVariant{Other: &other}})
	if err != nil {
		t.Fatal(err)
	}

	var value gotVariant
	if err := Unmarshal(encoded, []any{&value}); err == nil {
		t.Fatal("expected error for unknown selected variant arm")
	}
}

func BenchmarkUnmarshal_skippedVsMaterializedNestedFields(b *testing.B) {
	type ignoredInner struct {
		A     uint64   `ic:"a"`
		Maybe *uint32  `ic:"maybe"`
		Names []string `ic:"names"`
	}
	type wire struct {
		Ignored []ignoredInner `ic:"a"`
		Tail    string         `ic:"z"`
	}
	type got struct {
		Tail string `ic:"z"`
	}

	maybe := uint32(7)
	ignored := make([]ignoredInner, 64)
	for i := range ignored {
		ignored[i] = ignoredInner{
			A:     uint64(i),
			Maybe: &maybe,
			Names: []string{"alpha", "beta", "gamma"},
		}
	}
	encoded, err := Marshal([]any{wire{Ignored: ignored, Tail: "kept"}})
	if err != nil {
		b.Fatal(err)
	}

	b.Run("direct_skip", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var value got
			if err := Unmarshal(encoded, []any{&value}); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("materialize_any", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, _, err := Decode(encoded); err != nil {
				b.Fatal(err)
			}
		}
	})
}
