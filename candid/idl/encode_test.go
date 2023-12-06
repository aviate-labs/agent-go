package idl

import (
	"testing"
)

func TestEncode_issue7(t *testing.T) {
	type ConsumerPermissionEnum = struct {
		ReadOnly     *Null `ic:"ReadOnly,variant"`
		ReadAndWrite *Null `ic:"ReadAndWrite,variant"`
	}

	type SecretConsumer = struct {
		Name           string                 `ic:"name"`
		PermissionType ConsumerPermissionEnum `ic:"permission_type"`
	}

	raw, err := Marshal([]any{
		[]SecretConsumer{
			{
				Name:           "test",
				PermissionType: ConsumerPermissionEnum{ReadAndWrite: new(Null)},
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
