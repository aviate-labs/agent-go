package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleVariantType() {
	result := map[string]idl.Type{
		"ok":  new(idl.TextType),
		"err": new(idl.TextType),
	}
	typ := idl.NewVariantType(result)
	test_([]idl.Type{typ}, []any{idl.Variant{
		Name:  "ok",
		Value: "good",
		Type:  typ,
	}})
	test_([]idl.Type{idl.NewVariantType(result)}, []any{idl.Variant{
		Name:  "err",
		Value: "uhoh",
		Type:  typ,
	}})
	// Output:
	// 4449444c016b029cc20171e58eb4027101000004676f6f64
	// 4449444c016b029cc20171e58eb402710100010475686f68
}

func TestVariantType_UnmarshalGo(t *testing.T) {
	expectErr := func(t *testing.T, err error) {
		if err == nil {
			t.Fatal("expected error")
		} else {
			_, ok := err.(*idl.UnmarshalGoError)
			if !ok {
				t.Fatal("expected UnmarshalGoError")
			}
		}
	}
	t.Run("map", func(t *testing.T) {
		result := idl.VariantType{
			Fields: []idl.FieldType{
				{
					Name: "ok",
					Type: new(idl.TextType),
				},
				{
					Name: "err",
					Type: new(idl.TextType),
				},
			},
		}
		var m map[string]any
		if err := result.UnmarshalGo(map[string]any{"ok": "👌🏼"}, &m); err != nil {
			t.Fatal(err)
		}
		if err := result.UnmarshalGo(struct {
			Ok string `ic:"ok"`
		}{
			Ok: "👌🏼",
		}, &m); err != nil {
			t.Fatal(err)
		}
		if err := result.UnmarshalGo(map[string]string{
			"ok": "👌🏼",
		}, &m); err != nil {
			t.Fatal(err)
		}
		if m["ok"] != "👌🏼" {
			t.Fatal("expected 👌🏼")
		}
	})
	t.Run("struct", func(t *testing.T) {
		result := idl.VariantType{
			Fields: []idl.FieldType{
				{
					Name: "ok",
					Type: new(idl.TextType),
				},
				{
					Name: "err",
					Type: new(idl.TextType),
				},
			},
		}
		var m struct {
			Ok  *string
			Err *string
		}
		if err := result.UnmarshalGo(map[string]any{"ok": "👌🏼"}, &m); err != nil {
			t.Fatal(err)
		}
		if *m.Ok != "👌🏼" {
			t.Fatal("expected 👌🏼")
		}
		if err := result.UnmarshalGo(struct {
			Err string `ic:"err"`
		}{
			Err: "err",
		}, &m); err != nil {
			t.Fatal(err)
		}
		if *m.Err != "err" {
			t.Fatal("expected err")
		}
		ok := "🤔"
		if err := result.UnmarshalGo(struct {
			Ok *string `ic:"ok"`
		}{
			Ok: &ok,
		}, &m); err != nil {
			t.Fatal(err)
		}
		if *m.Ok != "🤔" {
			t.Fatal("expected 🤔")
		}
		if err := result.UnmarshalGo(map[string]string{
			"ok": "",
		}, &m); err != nil {
			t.Fatal(err)
		}
		if *m.Ok != "" {
			t.Fatal("expected empty string")
		}

		expectErr(t, result.UnmarshalGo(map[string]any{"ok": "👍🏼"}, &struct{ Ok string }{})) // Field must be a pointer.
		expectErr(t, result.UnmarshalGo(map[string]any{}, &m))                               // At least one field must be present.
		expectErr(t, result.UnmarshalGo(map[string]any{"ok": "👌🏼", "err": "👎🏼"}, &m))        // Only one field can be present.
	})

	var a any
	expectErr(t, (idl.VectorType{
		Type: idl.NullType{},
	}).UnmarshalGo(true, &a))
}
