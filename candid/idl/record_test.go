package idl_test

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"testing"
)

func ExampleRecordType() {
	test([]idl.Type{idl.NewRecordType(nil)}, []any{nil})
	test_([]idl.Type{idl.NewRecordType(map[string]idl.Type{
		"foo": new(idl.TextType),
		"bar": new(idl.IntType),
	})}, []any{
		map[string]any{
			"foo": "💩",
			"bar": idl.NewInt(42),
			"baz": idl.NewInt(0),
		},
	})
	// Output:
	// 4449444c016c000100
	// 4449444c016c02d3e3aa027c868eb7027101002a04f09f92a9
}

func ExampleRecordType_nested() {
	recordType := idl.NewRecordType(map[string]idl.Type{
		"foo": idl.Int32Type(),
		"bar": new(idl.BoolType),
	})
	recordValue := map[string]any{
		"foo": int32(42),
		"bar": true,
	}
	test_([]idl.Type{idl.NewRecordType(map[string]idl.Type{
		"foo": idl.Int32Type(),
		"bar": recordType,
		"baz": recordType,
		"bib": recordType,
	})}, []any{
		map[string]any{
			"foo": int32(42),
			"bar": recordValue,
			"baz": recordValue,
			"bib": recordValue,
		},
	})
	// Output:
	// 4449444c026c02d3e3aa027e868eb702756c04d3e3aa0200dbe3aa0200bbf1aa0200868eb702750101012a000000012a000000012a0000002a000000
}

func TestRecordType_UnmarshalGo(t *testing.T) {
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
		var m map[string]any
		if err := (idl.RecordType{}).UnmarshalGo(make(map[string]any), &m); err != nil {
			t.Fatal(err)
		}
		if err := (idl.RecordType{}).UnmarshalGo(struct{}{}, &m); err != nil {
			t.Fatal(err)
		}
		if err := (idl.RecordType{}).UnmarshalGo(make(map[string]idl.Nat), &m); err != nil {
			t.Fatal(err)
		}

		rt := idl.RecordType{
			Fields: []idl.FieldType{
				{
					Name: "foo",
					Type: new(idl.TextType),
				},
				{
					Name: "bar",
					Type: new(idl.IntType),
				},
			},
		}

		expectErr(t, rt.UnmarshalGo(make(map[string]idl.Null), &m))

		for range 3 {
			if err := rt.UnmarshalGo(map[string]any{
				"foo": "💩",
				"bar": idl.NewInt(42),
			}, &m); err != nil {
				t.Fatal(err)
			}
			if v, ok := m["foo"]; !ok || v != "💩" {
				t.Fatal(v)
			}
			if v, ok := m["bar"]; !ok || v.(idl.Int).BigInt().Int64() != 42 {
				t.Fatal(v)
			}
		}

		// Nested records.
		if err := (idl.RecordType{
			Fields: []idl.FieldType{
				{
					Name: "foo",
					Type: idl.RecordType{
						Fields: []idl.FieldType{
							{
								Name: "bar",
								Type: new(idl.IntType),
							},
						},
					},
				},
			},
		}).UnmarshalGo(map[string]any{
			"foo": map[string]any{
				"bar": idl.NewInt(42),
			},
		}, &m); err != nil {
			t.Fatal(err)
		}
		if v, ok := m["foo"]; !ok || v.(map[string]any)["bar"].(idl.Int).BigInt().Int64() != 42 {
			t.Fatal(v)
		}
	})
	t.Run("struct", func(t *testing.T) {
		var s struct {
			Foo string
			Bar idl.Int
		}
		if err := (idl.RecordType{}).UnmarshalGo(make(map[string]any), &s); err != nil {
			t.Fatal(err)
		}
		if err := (idl.RecordType{}).UnmarshalGo(struct{}{}, &s); err != nil {
			t.Fatal(err)
		}
		if err := (idl.RecordType{}).UnmarshalGo(make(map[string]idl.Nat), &s); err != nil {
			t.Fatal(err)
		}

		rt := idl.RecordType{
			Fields: []idl.FieldType{
				{
					Name: "foo",
					Type: new(idl.TextType),
				},
				{
					Name: "bar",
					Type: new(idl.IntType),
				},
			},
		}

		expectErr(t, rt.UnmarshalGo(make(map[string]idl.Null), &s))

		for range 3 {
			if err := rt.UnmarshalGo(map[string]any{
				"foo": "💩",
				"bar": idl.NewInt(42),
			}, &s); err != nil {
				t.Fatal(err)
			}
			if s.Foo != "💩" {
				t.Fatal(s)
			}
			if s.Bar.BigInt().Int64() != 42 {
				t.Fatal(s)
			}
		}

		expectErr(t, (idl.RecordType{
			Fields: []idl.FieldType{{Name: "unknown"}},
		}).UnmarshalGo(
			make(map[string]any), &rt,
		))

		rn := idl.RecordType{
			Fields: []idl.FieldType{
				{
					Name: "foo",
					Type: new(idl.TextType),
				},
				{
					Name: "bar",
					Type: idl.OptionalType{
						Type: idl.RecordType{},
					},
				},
			},
		}
		var sn struct {
			Foo string
			Bar *struct{}
		}
		if err := rn.UnmarshalGo(map[string]any{
			"foo": "💩",
			"bar": nil,
		}, &sn); err != nil {
			t.Fatal(err)
		}
	})

	var a any
	expectErr(t, (idl.VectorType{
		Type: idl.NullType{},
	}).UnmarshalGo(true, &a))
}
