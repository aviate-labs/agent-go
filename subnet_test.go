package agent

import (
	"testing"

	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/principal"
)

func TestAgent_GetSubnetMetrics(t *testing.T) {
	a, err := New(DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetSubnetMetrics(principal.MustDecode(certification.RootSubnetID)); err != nil {
		t.Fatal(err)
	}
}

func TestAgent_GetSubnets(t *testing.T) {
	a, err := New(DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetSubnets(); err != nil {
		t.Fatal(err)
	}
}

func TestAgent_GetSubnetsInfo(t *testing.T) {
	a, err := New(DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.GetSubnetsInfo(); err != nil {
		t.Fatal(err)
	}
}
