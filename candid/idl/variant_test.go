package idl_test

import "github.com/aviate-labs/agent-go/candid/idl"

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
