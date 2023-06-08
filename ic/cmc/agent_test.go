// Automatically generated by https://github.com/aviate-labs/agent-go.
package cmc_test

import (
	"github.com/aviate-labs/agent-go"

	"github.com/aviate-labs/agent-go/mock"
	"github.com/aviate-labs/agent-go/principal"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aviate-labs/agent-go/ic/cmc"
)

// Test_GetIcpXdrConversionRate tests the "get_icp_xdr_conversion_rate" method on the "cmc" canister.
func Test_GetIcpXdrConversionRate(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_icp_xdr_conversion_rate",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{*new(cmc.IcpXdrConversionRateResponse)}, nil
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

// Test_GetSubnetTypesToSubnets tests the "get_subnet_types_to_subnets" method on the "cmc" canister.
func Test_GetSubnetTypesToSubnets(t *testing.T) {
	a, err := newAgent([]mock.Method{
		{
			Name:      "get_subnet_types_to_subnets",
			Arguments: []any{},
			Handler: func(request mock.Request) ([]any, error) {
				return []any{*new(cmc.SubnetTypesToSubnetsResponse)}, nil
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
				return []any{*new(cmc.NotifyCreateCanisterResult)}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 cmc.NotifyCreateCanisterArg
	if _, err := a.NotifyCreateCanister(a0); err != nil {
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
				return []any{*new(cmc.NotifyTopUpResult)}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var a0 cmc.NotifyTopUpArg
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