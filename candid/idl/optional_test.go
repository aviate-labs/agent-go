package idl_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/idl"
)

func ExampleOpt() {
	var optNat = idl.NewOptionalType(new(idl.NatType))
	test([]idl.Type{optNat}, []any{nil})
	test([]idl.Type{optNat}, []any{idl.NewNat(uint(1))})
	// Output:
	// 4449444c016e7d010000
	// 4449444c016e7d01000101
}

func ExampleOpt_blob() {
	var optNatArray = idl.NewOptionalType(idl.VectorType{Type: idl.Nat8Type()})
	test([]idl.Type{optNatArray}, []any{nil})
	test([]idl.Type{optNatArray}, []any{[]byte{0x00}})
	// Output:
	// 4449444c026d7b6e00010100
	// 4449444c026d7b6e000101010100
}

func TestOptionalType_UnmarshalGo(t *testing.T) {
	if err := idl.UnmarshalGo(idl.OptionalType{
		Type: new(idl.NullType),
	}, nil, new(idl.Null)); err != nil {
		t.Fatal(err)
	}

	var nat *idl.Nat
	for range 3 {
		if err := idl.UnmarshalGo(idl.OptionalType{
			Type: new(idl.NatType),
		}, uint(1), &nat); err != nil {
			t.Fatal(err)
		}
		if nat == nil {
			t.Fatal("expected non-nil")
		}
		if (*nat).BigInt().Int64() != int64(1) {
			t.Fatal(nat)
		}
	}

	var a any
	if err := idl.UnmarshalGo(idl.OptionalType{
		Type: new(idl.NullType),
	}, "", &a); err == nil {
		t.Fatal("expected error")
	} else {
		var unmarshalGoError *idl.UnmarshalGoError
		if !errors.As(err, &unmarshalGoError) {
			t.Fatal("expected UnmarshalGoError")
		}
	}

	t.Run("Blob", func(t *testing.T) {
		var bs *[]byte
		if err := idl.UnmarshalGo(idl.OptionalType{
			Type: idl.NewVectorType(idl.Nat8Type()),
		}, []any{byte(0x00)}, &bs); err != nil {
			t.Error(err)
		}
	})
}

type motion struct {
	MotionText string `ic:"motion_text" json:"motion_text"`
}

type addedTag struct {
	Foo *uint64 `ic:"foo,omitempty" json:"foo,omitempty"`
}

type oldAction struct {
	Motion *motion `ic:"Motion,variant" json:"Motion,omitempty"`
}

type newAction struct {
	Motion   *motion   `ic:"Motion,variant" json:"Motion,omitempty"`
	AddedTag *addedTag `ic:"AddedTag,variant" json:"AddedTag,omitempty"`
}

type oldOptRecord struct {
	Action *oldAction `ic:"action,omitempty" json:"action,omitempty"`
}

type newOptRecord struct {
	Action *newAction `ic:"action,omitempty" json:"action,omitempty"`
}

type oldBareRecord struct {
	Action oldAction `ic:"action" json:"action"`
}

type newBareRecord struct {
	Action newAction `ic:"action" json:"action"`
}

// A tag both sides know must survive decoding into the older type.
func TestOptVariantKnownTag(t *testing.T) {
	bs, err := candid.Marshal([]any{newOptRecord{Action: &newAction{Motion: &motion{MotionText: "hello"}}}})
	if err != nil {
		t.Fatal(err)
	}
	var got oldOptRecord
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatal(err)
	}
	if got.Action == nil || got.Action.Motion == nil {
		t.Fatalf("known tag lost: %+v", got)
	}
	if got.Action.Motion.MotionText != "hello" {
		t.Fatalf("got %q", got.Action.Motion.MotionText)
	}
}

// Spec: a tag the receiver does not know, inside an opt, decodes to null.
// https://github.com/dfinity/candid/blob/master/spec/Candid.md
func TestOptVariantUnknownTagDecodesToNull(t *testing.T) {
	n := uint64(7)
	bs, err := candid.Marshal([]any{newOptRecord{Action: &newAction{AddedTag: &addedTag{Foo: &n}}}})
	if err != nil {
		t.Fatal(err)
	}
	var got oldOptRecord
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatalf("unknown tag in opt should decode to null, got error: %v", err)
	}
	if got.Action != nil {
		t.Fatalf("expected nil action, got %+v", got.Action)
	}
}

// Without the opt wrapper there is no null to fall back to, so an unknown
// tag must still be an error.
func TestBareVariantUnknownTagStillErrors(t *testing.T) {
	n := uint64(7)
	bs, err := candid.Marshal([]any{newBareRecord{Action: newAction{AddedTag: &addedTag{Foo: &n}}}})
	if err != nil {
		t.Fatal(err)
	}
	var got oldBareRecord
	if err := candid.Unmarshal(bs, []any{&got}); err == nil {
		t.Fatal("expected error for unknown tag in bare variant, got nil")
	}
}

// A record field the sender omits, present as opt in the receiver.
type recWithoutField struct {
	A uint64 `ic:"a" json:"a"`
}

type recWithOptField struct {
	A uint64  `ic:"a" json:"a"`
	B *uint64 `ic:"b,omitempty" json:"b,omitempty"`
}

func TestRecordMissingOptFieldIsNull(t *testing.T) {
	bs, err := candid.Marshal([]any{recWithoutField{A: 1}})
	if err != nil {
		t.Fatal(err)
	}
	var got recWithOptField
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatalf("missing opt field should be null, got error: %v", err)
	}
	if got.B != nil {
		t.Fatalf("expected nil B, got %v", *got.B)
	}
}

func TestRecordExtraFieldIgnored(t *testing.T) {
	n := uint64(9)
	bs, err := candid.Marshal([]any{recWithOptField{A: 1, B: &n}})
	if err != nil {
		t.Fatal(err)
	}
	var got recWithoutField
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatalf("extra field should be ignored, got error: %v", err)
	}
	if got.A != 1 {
		t.Fatalf("got A=%d", got.A)
	}
}

// "Any type is a subtype of an option": a changed constituent type under an
// opt is seen as null rather than being an error.
func TestFixOptTypeChangeDegradesToNull(t *testing.T) {
	type optNatRec struct {
		V *uint64 `ic:"v,omitempty" json:"v,omitempty"`
	}
	type optTextRec struct {
		V *string `ic:"v,omitempty" json:"v,omitempty"`
	}

	n := uint64(5)
	bs, err := candid.Marshal([]any{optNatRec{V: &n}})
	if err != nil {
		t.Fatal(err)
	}
	var got optTextRec
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatalf("opt type change should decode to null, got error: %v", err)
	}
	if got.V != nil {
		t.Fatalf("expected nil V, got %q", *got.V)
	}
}

// Degrading under nested opts must yield null, not an allocated outer pointer
// wrapping a nil inner one: the latter nil-derefs any caller that treats a
// non-nil outer as "value present".
func TestFixOptNestedTypeChangeDegradesToNull(t *testing.T) {
	type optOptNatRec struct {
		V **uint64 `ic:"v,omitempty" json:"v,omitempty"`
	}
	type optOptTextRec struct {
		V **string `ic:"v,omitempty" json:"v,omitempty"`
	}

	n := uint64(5)
	p := &n
	bs, err := candid.Marshal([]any{optOptNatRec{V: &p}})
	if err != nil {
		t.Fatal(err)
	}
	var got optOptTextRec
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatalf("nested opt type change should decode to null, got error: %v", err)
	}
	if got.V != nil {
		t.Fatalf("expected nil V, got outer non-nil (inner nil: %t)", *got.V == nil)
	}
}

// M(null : opt t) = i8(0), M(?v : opt t) = i8(1) M(v : t), so at opt opt nat
// `opt null` is 0x01 0x00 and `null` is 0x00. Encoding both as 0x00 loses a
// distinction the spec makes.
func TestOptNullEncodesDistinctlyFromNull(t *testing.T) {
	type optOptNatRec struct {
		V **uint64 `ic:"v,omitempty" json:"v,omitempty"`
	}

	var inner *uint64
	optNull, err := candid.Marshal([]any{optOptNatRec{V: &inner}})
	if err != nil {
		t.Fatal(err)
	}
	null, err := candid.Marshal([]any{optOptNatRec{V: nil}})
	if err != nil {
		t.Fatal(err)
	}
	if string(optNull) == string(null) {
		t.Fatalf("opt(null) and null encode identically: %x", optNull)
	}
}

// The counterpart to the degrade case: an inner null the sender really did
// transmit must survive as a non-nil outer pointer, not collapse to null.
func TestOptNullRoundTrips(t *testing.T) {
	type optOptNatRec struct {
		V **uint64 `ic:"v,omitempty" json:"v,omitempty"`
	}

	var inner *uint64
	bs, err := candid.Marshal([]any{optOptNatRec{V: &inner}})
	if err != nil {
		t.Fatal(err)
	}
	var got optOptNatRec
	if err := candid.Unmarshal(bs, []any{&got}); err != nil {
		t.Fatal(err)
	}
	if got.V == nil {
		t.Fatal("opt(null) collapsed to null")
	}
	if *got.V != nil {
		t.Fatalf("expected inner nil, got %d", **got.V)
	}
}

// Value bytes from the "nested opt" block of dfinity/candid
// test/construct.test.did, at type (opt opt bool).
func TestNestedOptUpstreamVectors(t *testing.T) {
	oob := idl.NewOptionalType(idl.NewOptionalType(new(idl.BoolType)))
	f := false
	pf := &f
	var pnil *bool

	for _, tc := range []struct {
		name string
		val  any
		enc  []byte
		raw  any
	}{
		{"null", nil, []byte{0x00}, nil},
		{"opt null", &pnil, []byte{0x01, 0x00}, idl.Some{}},
		{"opt opt false", &pf, []byte{0x01, 0x01, 0x00}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bs, err := oob.EncodeValue(tc.val)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(bs, tc.enc) {
				t.Errorf("encode = %x, want %x", bs, tc.enc)
			}
			raw, err := oob.Decode(bytes.NewReader(tc.enc))
			if err != nil {
				t.Fatal(err)
			}
			if raw != tc.raw {
				t.Errorf("decode = %#v, want %#v", raw, tc.raw)
			}
		})
	}
}
