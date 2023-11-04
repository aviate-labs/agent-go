package idl_test

import (
	"fmt"

	"github.com/aviate-labs/agent-go/candid/idl"
)

func Example_ledger() {
	fmt.Println(idl.NewInterface(func(typ idl.IDL) *idl.Service {
		accountIdentitier := typ.Vec(typ.Nat8)
		accountBalanceArgs := typ.Record(map[string]idl.Type{
			"account": accountIdentitier,
		})
		tokens := idl.NewRecordType(map[string]idl.Type{
			"e8s": idl.Nat64Type(),
		})
		return typ.Service(map[string]*idl.FunctionType{
			"account_balance": typ.Func([]idl.FunctionParameter{{Type: accountBalanceArgs}}, []idl.FunctionParameter{{Type: tokens}}, []string{"query"}),
			// etc.
		})
	}))
	// Output:
	// service {account_balance:(record {account:vec nat8}) -> (record {e8s:nat64}) query}
}

func Example_optionalNat() {
	fmt.Println(idl.NewInterface(func(typ idl.IDL) *idl.Service {
		time := idl.NewOptionalType(new(idl.NatType))
		return typ.Service(map[string]*idl.FunctionType{
			"now": typ.Func([]idl.FunctionParameter{}, []idl.FunctionParameter{{Type: time}}, []string{"query"}),
			// etc.
		})
	}))
	// Output:
	// service {now:() -> (opt nat) query}
}

func Example_tokens() {
	fmt.Println(idl.NewRecordType(map[string]idl.Type{
		"e8s": idl.Nat64Type(),
	}))
	// Output:
	// record {e8s:nat64}
}
