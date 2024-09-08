package wallet

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
)

// replace with "dfx identity export xxx" result
var pem = []byte(`
-----BEGIN EC PRIVATE KEY-----
-----END EC PRIVATE KEY-----`)

func Test_WalletCanister_LocalNet(t *testing.T) {
	id, err := identity.NewSecp256k1IdentityFromPEMWithoutParameters(pem)
	if err != nil {
		panic(err)
	}

	host, err := url.Parse("http://localhost:4943")
	if err != nil {
		panic(err)
	}
	cfg := agent.Config{
		Identity:                       id,
		ClientConfig:                   &agent.ClientConfig{Host: host},
		FetchRootKey:                   true,
		PollTimeout:                    30 * time.Second,
		DisableSignedQueryVerification: true, //MUST BE TRUE TO ACCESS LOCAL REPLICA
	}
	a, err := NewAgent(principal.MustDecode("bnz7o-iuaaa-aaaaa-qaaaa-cai"), cfg)
	if err != nil {
		panic(err)
	}

	balance, err := a.WalletBalance()
	if err != nil {
		panic(err)
	}
	fmt.Printf("balance:%v\n", balance)

	canisterId := principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai")

	var s1 string
	err = a.Query(canisterId, "greet", []any{}, []any{&s1})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s1:%v\n", s1)

	var s2 string
	err = a.Query(canisterId, "concat", []any{"hello", "world"}, []any{&s2})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s2:%v\n", s2)

	var s3 string
	err = a.Call(canisterId, "sha256", []any{"hello, world", uint32(2)}, []any{&s3}) //2's type should match with taht defined in hasher canister.
	if err != nil {
		panic(err)
	}
	fmt.Printf("s3:%v\n", s3)

	//step4: concat_with_cycles
	input4, err := idl.Marshal([]any{"hello", "world"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("input4:%v\n", input4)

	arg4 := struct {
		Canister   principal.Principal `ic:"canister" json:"canister"`
		MethodName string              `ic:"method_name" json:"method_name"`
		Args       []byte              `ic:"args" json:"args"`
		Cycles     uint64              `ic:"cycles" json:"cycles"`
	}{
		Canister:   principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai"),
		MethodName: "concat_with_cycles",
		Args:       []byte(input4),
		Cycles:     200_000_000,
	}
	res4, err := a.WalletCall(arg4)
	if err != nil {
		panic(err)
	}
	var s4 string
	err = idl.Unmarshal(res4.Ok.Return, []any{&s4})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s4:%v\n", s4)

	//step5, sha256_with_cycles
	n := uint32(2)
	input5, err := idl.Marshal([]any{"hello, world", n})
	if err != nil {
		panic(err)
	}
	fmt.Printf("input5:%v\n", input5)

	arg5 := struct {
		Canister   principal.Principal `ic:"canister" json:"canister"`
		MethodName string              `ic:"method_name" json:"method_name"`
		Args       []byte              `ic:"args" json:"args"`
		Cycles     uint64              `ic:"cycles" json:"cycles"`
	}{
		Canister:   principal.MustDecode("bkyz2-fmaaa-aaaaa-qaaaq-cai"),
		MethodName: "sha256_with_cycles",
		Args:       []byte(input5),
		Cycles:     uint64(200_000_000 * n),
	}
	res5, err := a.WalletCall(arg5)
	if err != nil {
		panic(err)
	}

	var s5 string
	err = idl.Unmarshal(res5.Ok.Return, []any{&s5})
	if err != nil {
		panic(err)
	}
	fmt.Printf("s5:%v\n", s5)
}
