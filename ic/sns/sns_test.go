package sns

import (
	"bytes"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/sns/root"
	"testing"
)

func TestSNS(t *testing.T) {
	snsAgent, err := NewAgent(ic.SNS_WASM_PRINCIPAL, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	snsList, err := snsAgent.ListDeployedSnses(struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	sns1Instances := snsList.Instances[0]
	rootAgent, err := root.NewAgent(*sns1Instances.RootCanisterId, agent.DefaultConfig)
	if err != nil {
		t.Fatal(err)
	}
	canisterStatus, err := rootAgent.CanisterStatus(root.CanisterIdRecord{
		CanisterId: *sns1Instances.RootCanisterId,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(canisterStatus.Settings.Controllers) != 1 || !bytes.Equal(canisterStatus.Settings.Controllers[0].Raw, sns1Instances.GovernanceCanisterId.Raw) {
		t.Fatalf("unexpected controllers: %v", canisterStatus.Settings.Controllers)
	}
}
