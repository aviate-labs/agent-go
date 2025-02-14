package candid

import (
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
