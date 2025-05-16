package candid

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func TestEncode_issue7(t *testing.T) {
	type ConsumerPermissionEnum = struct {
		ReadOnly     *idl.Null `ic:"ReadOnly,variant"`
		ReadAndWrite *idl.Null `ic:"ReadAndWrite,variant"`
	}

	type SecretConsumer = struct {
		Name           string                 `ic:"name"`
		PermissionType ConsumerPermissionEnum `ic:"permission_type"`
	}

	raw, err := Marshal([]any{
		[]SecretConsumer{
			{
				Name:           "test",
				PermissionType: ConsumerPermissionEnum{ReadAndWrite: new(idl.Null)},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var v []SecretConsumer
	if err := Unmarshal(raw, []any{&v}); err != nil {
		t.Fatal(err)
	}
}

func TestRecordTupleFields(t *testing.T) {
	type T struct {
		Field0 string `ic:"0,tuple"`
		Field1 string `ic:"1,tuple"`
	}
	raw, err := Marshal([]any{
		T{
			Field0: "hello",
			Field1: "world",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := hex.DecodeString("4449444c016c020071017101000568656c6c6f05776f726c64")
	if !bytes.Equal(raw, expected) {
		t.Fatalf("expected %x, got %x", expected, raw)
	}
	var v T
	if err := Unmarshal(expected, []any{&v}); err != nil {
		t.Fatal(err)
	}
	if v.Field0 != "hello" {
		t.Fatalf("expected hello, got %s", v.Field0)
	}
	if v.Field1 != "world" {
		t.Fatalf("expected world, got %s", v.Field1)
	}
}

func TestVariantType_default(t *testing.T) {
	type V = struct {
		A *idl.Null `ic:"A,variant"`
		B *idl.Null `ic:"B,variant"`
	}
	raw, err := Marshal([]any{
		V{},
	})
	if err == nil {
		t.Error("expected error")
	}
	var v V
	if err := Unmarshal(raw, []any{&v}); err == nil {
		t.Error("expected error")
	}
}
