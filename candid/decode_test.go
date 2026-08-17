package candid

import (
	"bytes"
	"math/big"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
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

// An argument the sender omits entirely is null when the receiver expects an
// option, per the "missing argument" vectors in the upstream test suite.
func TestUnmarshalMissingTrailingOptArg(t *testing.T) {
	empty := []byte("DIDL\x00\x00")

	var p *uint64
	if err := Unmarshal(empty, []any{&p}); err != nil {
		t.Fatalf("missing opt arg should decode to null, got: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil, got %d", *p)
	}

	// A missing non-optional argument is still an error.
	var n uint64
	if err := Unmarshal(empty, []any{&n}); err == nil {
		t.Fatal("expected error for missing nat argument")
	}
}

// dfinity/candid construct.test.did: an uninhabited recursive type must be
// rejected, not decoded. Recursing on it overflows the stack, which Go reports
// as a fatal error that no caller can recover from.
func TestDecodeRecursiveRecordTerminates(t *testing.T) {
	// type EmptyRecord = record { EmptyRecord }
	payload := []byte{'D', 'I', 'D', 'L', 0x01, 0x6c, 0x01, 0x00, 0x00, 0x01, 0x00}
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		_, _, _ = Decode(payload)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("decode did not terminate")
	}
}

// dfinity/candid construct.test.did: `opt` whose element is itself.
func TestDecodeSelfReferentialOpt(t *testing.T) {
	// DIDL, 1 type: opt(index 0), 0 args.
	payload := []byte{'D', 'I', 'D', 'L', 0x01, 0x6e, 0x00, 0x00}
	if _, _, err := Decode(payload); err != nil {
		t.Fatalf("self-referential opt: %v", err)
	}
}

// dfinity/candid construct.test.did: a vector of a zero-width element type
// carries no payload bytes, so its length is not bounded by bytes remaining.
func TestDecodeVecOfZeroWidth(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"vec null", []byte{'D', 'I', 'D', 'L', 0x01, 0x6d, 0x7f, 0x01, 0x00, 0xe8, 0x07}},
		{"vec record {}", []byte{'D', 'I', 'D', 'L', 0x02, 0x6d, 0x01, 0x6c, 0x00, 0x01, 0x00, 0x05}},
		{"vec reserved", []byte{'D', 'I', 'D', 'L', 0x01, 0x6d, 0x70, 0x01, 0x00, 0x01}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A length no payload could justify is refused rather than allocated.
func TestDecodeVecLengthBomb(t *testing.T) {
	// vec null, length 1_000_000_000.
	payload := []byte{'D', 'I', 'D', 'L', 0x01, 0x6d, 0x7f, 0x01, 0x00, 0x80, 0x94, 0xeb, 0xdc, 0x03}
	if _, _, err := Decode(payload); err == nil {
		t.Fatal("expected length bomb to be refused")
	}
}

// dfinity/candid reference.test.did: service method names must be unique,
// sorted, and valid utf8.
func TestDecodeServiceMethodValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"unsorted", []byte("DIDL\x02\x6a\x01\x71\x01\x7d\x00\x69\x02\x04foo2\x00\x03foo\x00\x01\x01\x01\x03\xca\xff\xee")},
		{"duplicate", []byte("DIDL\x02\x6a\x01\x71\x01\x7d\x00\x69\x02\x03foo\x00\x03foo\x00\x01\x01\x01\x03\xca\xff\xee")},
		{"invalid utf8", []byte("DIDL\x02\x6a\x01\x71\x01\x7d\x00\x69\x01\x03\xe2\x28\xa1\x00\x01\x01\x01\x03\xca\xff\xee")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err == nil {
				t.Fatal("expected decode failure")
			}
		})
	}
}

// dfinity/candid reference.test.did: a principal whose declared length exceeds
// the remaining bytes must be rejected.
func TestDecodePrincipalLength(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"too short", []byte("DIDL\x00\x01\x68\x01\x03\xca\xff")},
		{"length beyond payload", []byte("DIDL\x00\x01\x68\x01\x80\x94\xeb\xdc\x03Motoko")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err == nil {
				t.Fatal("expected decode failure")
			}
		})
	}
}

// A length prefix larger than the remaining input must fail, not silently
// yield a short value. r.Read reports a short count without an error.
func TestDecodeTruncatedLengths(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		// Future type: 4 bytes declared, 1 present.
		{"future type body", []byte("DIDL\x01\x80\x01\x04\xaa\x01\x00")},
		// Func annotation count beyond the payload.
		{"func annotations", []byte("DIDL\x01\x6a\x00\x00\x7f\x01\x00")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err == nil {
				t.Fatal("expected decode failure")
			}
		})
	}
}

// dfinity/candid spacebomb.test.did: a record holding two copies of the one
// below it doubles per level, so a short type table names millions of values.
func TestDecodeTypeSizeBomb(t *testing.T) {
	payload := []byte("DIDL\x17\x6c\x02\x01\x7f\x02\x7f\x6c\x02\x01\x00\x02\x00\x6c\x02\x00\x01\x01\x01\x6c\x02\x00\x02\x01\x02\x6c\x02\x00\x03\x01\x03\x6c\x02\x00\x04\x01\x04\x6c\x02\x00\x05\x01\x05\x6c\x02\x00\x06\x01\x06\x6c\x02\x00\x07\x01\x07\x6c\x02\x00\x08\x01\x08\x6c\x02\x00\x09\x01\x09\x6c\x02\x00\x0a\x01\x0a\x6c\x02\x00\x0b\x01\x0b\x6c\x02\x00\x0c\x01\x0c\x6c\x02\x00\x0d\x02\x0d\x6c\x02\x00\x0e\x01\x0e\x6c\x02\x00\x0f\x01\x0f\x6c\x02\x00\x10\x01\x10\x6c\x02\x00\x11\x01\x11\x6c\x02\x00\x12\x01\x12\x6c\x02\x00\x13\x01\x13\x6e\x14\x6d\x15\x01\x16\x02\x01\x01")
	if _, _, err := Decode(payload); err == nil {
		t.Fatal("expected size bomb to be refused")
	}
}

// Nested zero-width vectors: each length is plausible, the product is not.
func TestDecodeNestedZeroWidthVec(t *testing.T) {
	payload := []byte("DIDL\x02\x6d\x01\x6d\x7f\x01\x00\x05\xff\xff\x3f\xff\xff\x3f\xff\xff\x3f\xff\xff\x3f\xff\xff\x3f")
	if _, _, err := Decode(payload); err == nil {
		t.Fatal("expected nested length bomb to be refused")
	}
}

// A cycle running through a func must not restart the String() walk: the walk
// is unbounded, and a stack overflow is a fatal error that no caller can
// recover from, so it only shows up in a subprocess.
func TestDecodeCyclicFuncString(t *testing.T) {
	// 0: record { 0 : 1 }, 1: func (0) -> (); one arg of type 0.
	payload := []byte("DIDL\x02\x6c\x01\x00\x01\x6a\x01\x00\x00\x00\x01\x00\x01\x01\x01\xaa\x01m")
	if os.Getenv("CYCLIC_FUNC_CHILD") == "1" {
		ts, _, err := Decode(payload)
		if err != nil {
			return // rejecting the type outright is also fine
		}
		for _, t := range ts {
			_ = t.String()
		}
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestDecodeCyclicFuncString$", "-test.timeout=30s")
	cmd.Env = append(os.Environ(), "CYCLIC_FUNC_CHILD=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("String() on a cyclic func type: %v\n%s", err, out)
	}
}

// An outer length of 1 must not hand its whole budget to the child: dividing by
// the element count leaves the cap unchanged and the nesting still multiplies.
func TestDecodeNestedZeroWidthVecSingleOuter(t *testing.T) {
	// vec (vec null), outer length 1, inner length 1<<20.
	payload := []byte("DIDL\x02\x6d\x01\x6d\x7f\x01\x00\x01\x80\x80\x40")
	if _, _, err := Decode(payload); err == nil {
		t.Fatal("expected nested length bomb to be refused")
	}
}

// The budget must survive an opt or record on the path: those call the plain
// Decode entry point, which would reset the cap to the full budget.
func TestDecodeZeroWidthVecBehindOptAndRecord(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		// vec (opt (vec null)): outer 1, opt present, inner 1<<20.
		{"opt", []byte("DIDL\x03\x6d\x01\x6e\x02\x6d\x7f\x01\x00\x01\x01\x80\x80\x40")},
		// vec (record { 0 : vec null }): outer 1, inner 1<<20.
		{"record", []byte("DIDL\x03\x6d\x01\x6c\x01\x00\x02\x6d\x7f\x01\x00\x01\x80\x80\x40")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err == nil {
				t.Fatal("expected nested length bomb to be refused")
			}
		})
	}
}

// A service whose signatures fail to resolve must not keep a partially filled
// method list: String and AddTypeDefinition both dereference Func.
func TestDecodeServicePartialResolve(t *testing.T) {
	// Method 1 points at a non-func type, so resolution fails after method 0.
	payload := []byte("DIDL\x02\x6a\x00\x00\x00\x69\x02\x01a\x00\x01b\x01\x00")
	ts, _, err := Decode(payload)
	if err != nil {
		return
	}
	for _, typ := range ts {
		s, ok := typ.(*idl.ServiceType)
		if !ok {
			continue
		}
		for _, m := range s.Methods {
			if m.Func == nil {
				t.Fatalf("method %s left unresolved", m.Name)
			}
		}
		_ = s.String()
	}
}

// isEmptyType has to see through the RecursiveType that resolve() installs,
// otherwise an uninhabited record escapes the EmptyType substitution.
func TestIsEmptyTypeThroughRecursive(t *testing.T) {
	rec := idl.NewRecursiveType("rec_0")
	rec.SetInner(idl.NewRecordType(map[string]idl.Type{"a": rec}))
	if !isEmptyType(rec, nil) {
		t.Fatal("record cycle through a RecursiveType not reported as empty")
	}
}

// An unresolved placeholder must report an error rather than dereference nil.
func TestRecursiveDecodeUnresolved(t *testing.T) {
	rec := idl.NewRecursiveType("rec_0")
	if _, err := rec.Decode(bytes.NewReader([]byte{0x00}), nil); err == nil {
		t.Fatal("expected an error decoding an unresolved recursive type")
	}
}

// The quota is settable per call in both directions: tighter than the default,
// and off entirely via a nil budget.
func TestDecodeQuota(t *testing.T) {
	vs := make([]any, 20000)
	for i := range vs {
		vs[i] = idl.NewNat(uint64(i))
	}
	payload, err := Marshal([]any{vs})
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := Decode(payload); err != nil {
		t.Fatalf("default quota: %v", err)
	}
	if _, _, err := DecodeWithQuota(payload, idl.NewBudget(1000)); err == nil {
		t.Fatal("expected a tight quota to refuse the payload")
	}
	if _, _, err := DecodeWithQuota(payload, idl.NewUnlimitedBudget()); err != nil {
		t.Fatalf("unmetered: %v", err)
	}

	// A bomb the default refuses still decodes unmetered, so the budget is
	// really in force rather than silently ignored.
	bomb := []byte{'D', 'I', 'D', 'L', 0x01, 0x6d, 0x7f, 0x01, 0x00, 0x80, 0x89, 0x7a}
	if _, _, err := Decode(bomb); err == nil {
		t.Fatal("expected the default quota to refuse the bomb")
	}
	if _, _, err := DecodeWithQuota(bomb, idl.NewUnlimitedBudget()); err != nil {
		t.Fatalf("unmetered large vec: %v", err)
	}

	var out []idl.Nat
	if err := UnmarshalWithQuota(payload, []any{&out}, idl.NewBudget(1000)); err == nil {
		t.Fatal("expected Unmarshal to honour a tight quota")
	}
	if err := Unmarshal(payload, []any{&out}); err != nil {
		t.Fatalf("Unmarshal default quota: %v", err)
	}
}

// A declared length with no bytes behind it must be refused before the
// allocation: io.ReadFull below each make() catches the short read, but only
// after the memory is already committed.
func TestDecodeUnboundedTypeTableLengths(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{"service method name", []byte("DIDL\x01\x69\x01\x80\x80\x80\x80\x80\x02\x00\x00")},
		{"func annotations", []byte("DIDL\x01\x6a\x00\x00\x80\x80\x80\x80\x80\x02\x00\x00")},
		{"future type body", []byte("DIDL\x01\x67\x80\x80\x80\x80\x80\x02\x00\x00")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := Decode(tc.payload); err == nil {
				t.Fatal("expected decode failure")
			}
		})
	}
}
