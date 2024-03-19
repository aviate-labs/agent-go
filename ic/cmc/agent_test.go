// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
package cmc_test

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/mock"
	"github.com/aviate-labs/agent-go/principal"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aviate-labs/agent-go/ic/cmc"
)

// Test_CreateCanister tests the "create_canister" method on the "cmc" canister.
func Test_CreateCanister(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "create_canister",
			Arguments: []any{new(cmc.CreateCanisterArg)},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.CreateCanisterResult{
					Ok: new(principal.Principal),
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 = cmc.CreateCanisterArg{
		*new(*cmc.CanisterSettings),
		*new(*string),
		*new(*cmc.SubnetSelection),
	}
	if _, err := a.CreateCanister(a0); err != nil {
		t.Fatal(err)
	}

}

// Test_GetBuildMetadata tests the "get_build_metadata" method on the "cmc" canister.
func Test_GetBuildMetadata(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_build_metadata",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{*new(string)}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.GetBuildMetadata(); err != nil {
		t.Fatal(err)
	}

}

// Test_GetIcpXdrConversionRate tests the "get_icp_xdr_conversion_rate" method on the "cmc" canister.
func Test_GetIcpXdrConversionRate(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_icp_xdr_conversion_rate",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.IcpXdrConversionRateResponse{
					cmc.IcpXdrConversionRate{
						*new(uint64),
						*new(uint64),
					},
					*new([]byte),
					*new([]byte),
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.GetIcpXdrConversionRate(); err != nil {
		t.Fatal(err)
	}

}

// Test_GetPrincipalsAuthorizedToCreateCanistersToSubnets tests the "get_principals_authorized_to_create_canisters_to_subnets" method on the "cmc" canister.
func Test_GetPrincipalsAuthorizedToCreateCanistersToSubnets(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_principals_authorized_to_create_canisters_to_subnets",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.PrincipalsAuthorizedToCreateCanistersToSubnetsResponse{
					[]struct {
						Field0 principal.Principal   `ic:"0" json:"0"`
						Field1 []principal.Principal `ic:"1" json:"1"`
					}{

						{
							*new(principal.Principal),
							[]principal.Principal{*new(principal.Principal)},
						}},
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.GetPrincipalsAuthorizedToCreateCanistersToSubnets(); err != nil {
		t.Fatal(err)
	}

}

// Test_GetSubnetTypesToSubnets tests the "get_subnet_types_to_subnets" method on the "cmc" canister.
func Test_GetSubnetTypesToSubnets(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_subnet_types_to_subnets",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.SubnetTypesToSubnetsResponse{
					[]struct {
						Field0 string                `ic:"0" json:"0"`
						Field1 []principal.Principal `ic:"1" json:"1"`
					}{

						{
							*new(string),
							[]principal.Principal{*new(principal.Principal)},
						}},
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.GetSubnetTypesToSubnets(); err != nil {
		t.Fatal(err)
	}

}

// Test_NotifyCreateCanister tests the "notify_create_canister" method on the "cmc" canister.
func Test_NotifyCreateCanister(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "notify_create_canister",
			Arguments: []any{new(cmc.NotifyCreateCanisterArg)},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.NotifyCreateCanisterResult{
					Ok: new(principal.Principal),
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 = cmc.NotifyCreateCanisterArg{
		*new(uint64),
		*new(principal.Principal),
		*new(*string),
		*new(*cmc.SubnetSelection),
		*new(*cmc.CanisterSettings),
	}
	if _, err := a.NotifyCreateCanister(a0); err != nil {
		t.Fatal(err)
	}

}

// Test_NotifyMintCycles tests the "notify_mint_cycles" method on the "cmc" canister.
func Test_NotifyMintCycles(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "notify_mint_cycles",
			Arguments: []any{new(cmc.NotifyMintCyclesArg)},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.NotifyMintCyclesResult{
					Ok: idl.Ptr(cmc.NotifyMintCyclesSuccess{
						idl.NewNat(uint(0)),
						idl.NewNat(uint(0)),
						idl.NewNat(uint(0)),
					}),
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 = cmc.NotifyMintCyclesArg{
		*new(uint64),
		*new(*[]byte),
		*new(*[]byte),
	}
	if _, err := a.NotifyMintCycles(a0); err != nil {
		t.Fatal(err)
	}

}

// Test_NotifyTopUp tests the "notify_top_up" method on the "cmc" canister.
func Test_NotifyTopUp(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "notify_top_up",
			Arguments: []any{new(cmc.NotifyTopUpArg)},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{cmc.NotifyTopUpResult{
					Ok: idl.Ptr(idl.NewNat(uint(0))),
				}}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 = cmc.NotifyTopUpArg{
		*new(uint64),
		*new(principal.Principal),
	}
	if _, err := a.NotifyTopUp(a0); err != nil {
		t.Fatal(err)
	}

}

// newAgent creates a new agent with the given (mock) methods.
// Runs a mock replica in the background.
func newAgent(methods []mock.Method) (*cmc.Agent, error) {
	replica := mock.NewReplica()
	canisterId := principal.Principal{Raw: []byte("cmc")}
	replica.AddCanister(canisterId, methods)
	s := httptest.NewServer(replica)
	u, _ := url.Parse(s.URL)
	a, err := cmc.NewAgent(canisterId, agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}
