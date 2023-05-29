package ic_test

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/mock"
	"github.com/aviate-labs/agent-go/principal"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestTypesAgent_prim(t *testing.T) {
	replica := mock.NewReplica()
	var canisterId principal.Principal
	replica.AddCanister(
		canisterId,
		[]mock.Method{
			{
				Name:      "nat",
				Arguments: []any{new(idl.Nat), new(uint8), new(uint16), new(uint32), new(uint64)},
				Handler:   passArguments,
			},
			{
				Name:      "vec_nat",
				Arguments: []any{new([]idl.Nat), new([]uint8), new([]uint16), new([]uint32), new([]uint64)},
				Handler:   passArguments,
			},
			{
				Name:      "int",
				Arguments: []any{new(idl.Int), new(int8), new(int16), new(int32), new(int64)},
				Handler:   passArguments,
			},
			{
				Name:      "vec_int",
				Arguments: []any{new([]idl.Int), new([]int8), new([]int16), new([]int32), new([]int64)},
				Handler:   passArguments,
			},
			{
				Name:      "float",
				Arguments: []any{new(float32), new([]float32), new(float64), new([]float64)},
				Handler:   passArguments,
			},
			{
				Name:      "text",
				Arguments: []any{new(string), new([]string)},
				Handler:   passArguments,
			},
			{
				Name:      "bool",
				Arguments: []any{new(bool), new([]bool)},
				Handler:   passArguments,
			},
			{
				Name:      "principal",
				Arguments: []any{new(principal.Principal), new([]principal.Principal)},
				Handler:   passArguments,
			},
		},
	)

	s := httptest.NewServer(replica)
	u, _ := url.Parse(s.URL)
	a, _ := NewTypesAgent(canisterId, agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	})

	t.Run("nat", func(t *testing.T) {
		var a0 = idl.NewNat(uint(0xFF))
		var a1 = uint8(8)
		var a2 = uint16(16)
		var a3 = uint32(32)
		var a4 = uint64(64)
		r0, r1, r2, r3, r4, err := a.Nat(a0, a1, a2, a3, a4)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a0, *r0) {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if a1 != *r1 {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
		if a2 != *r2 {
			t.Errorf("expected %v, got %v", a2, *r2)
		}
		if a3 != *r3 {
			t.Errorf("expected %v, got %v", a3, *r3)
		}
		if a4 != *r4 {
			t.Errorf("expected %v, got %v", a4, *r4)
		}
	})
	t.Run("vec nat", func(t *testing.T) {
		var a0 = []idl.Nat{idl.NewNat(uint(0xFF))}
		var a1 = []uint8{8}
		var a2 = []uint16{16}
		var a3 = []uint32{32}
		var a4 = []uint64{64}
		r0, r1, r2, r3, r4, err := a.VecNat(a0, a1, a2, a3, a4)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a0, *r0) {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if !reflect.DeepEqual(a1, *r1) {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
		if !reflect.DeepEqual(a2, *r2) {
			t.Errorf("expected %v, got %v", a2, *r2)
		}
		if !reflect.DeepEqual(a3, *r3) {
			t.Errorf("expected %v, got %v", a3, *r3)
		}
		if !reflect.DeepEqual(a4, *r4) {
			t.Errorf("expected %v, got %v", a4, *r4)
		}
	})
	t.Run("int", func(t *testing.T) {
		var a0 = idl.NewInt(0xFF)
		var a1 = int8(8)
		var a2 = int16(16)
		var a3 = int32(32)
		var a4 = int64(64)
		r0, r1, r2, r3, r4, err := a.Int(a0, a1, a2, a3, a4)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a0, *r0) {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if a1 != *r1 {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
		if a2 != *r2 {
			t.Errorf("expected %v, got %v", a2, *r2)
		}
		if a3 != *r3 {
			t.Errorf("expected %v, got %v", a3, *r3)
		}
		if a4 != *r4 {
			t.Errorf("expected %v, got %v", a4, *r4)
		}
	})
	t.Run("vec int", func(t *testing.T) {
		var a0 = []idl.Int{idl.NewInt(0xFF)}
		var a1 = []int8{8}
		var a2 = []int16{16}
		var a3 = []int32{32}
		var a4 = []int64{64}
		r0, r1, r2, r3, r4, err := a.VecInt(a0, a1, a2, a3, a4)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a0, *r0) {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if !reflect.DeepEqual(a1, *r1) {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
		if !reflect.DeepEqual(a2, *r2) {
			t.Errorf("expected %v, got %v", a2, *r2)
		}
		if !reflect.DeepEqual(a3, *r3) {
			t.Errorf("expected %v, got %v", a3, *r3)
		}
		if !reflect.DeepEqual(a4, *r4) {
			t.Errorf("expected %v, got %v", a4, *r4)
		}
	})
	t.Run("float", func(t *testing.T) {
		var a0 = float32(3.2)
		var a1 = []float32{3.2}
		var a2 = float64(6.4)
		var a3 = []float64{6.4}
		r0, r1, r2, r3, err := a.Float(a0, a1, a2, a3)
		if err != nil {
			t.Fatal(err)
		}
		if a0 != *r0 {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if !reflect.DeepEqual(a1, *r1) {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
		if a2 != *r2 {
			t.Errorf("expected %v, got %v", a2, *r2)
		}
		if !reflect.DeepEqual(a3, *r3) {
			t.Errorf("expected %v, got %v", a3, *r3)
		}
	})
	t.Run("text", func(t *testing.T) {
		var a0 = "a0"
		var a1 = []string{"a1"}
		r0, r1, err := a.Text(a0, a1)
		if err != nil {
			t.Fatal(err)
		}
		if a0 != *r0 {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if !reflect.DeepEqual(a1, *r1) {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
	})
	t.Run("bool", func(t *testing.T) {
		var a0 = true
		var a1 = []bool{true}
		r0, r1, err := a.Bool(a0, a1)
		if err != nil {
			t.Fatal(err)
		}
		if a0 != *r0 {
			t.Errorf("expected %v, got %v", a0, *r0)
		}
		if !reflect.DeepEqual(a1, *r1) {
			t.Errorf("expected %v, got %v", a1, *r1)
		}
	})
}

func passArguments(request mock.Request) ([]any, error) {
	return removePtr(request.Arguments), nil
}

func removePtr(args []any) []any {
	var result []any
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		result = append(result, v.Elem().Interface())
	}
	return result
}
