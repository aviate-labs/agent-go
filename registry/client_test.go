package registry

import (
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
)

var client, _ = New()

func TestClient_GetNodeList(t *testing.T) {
	nodes, err := client.GetNodeList()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) == 0 {
		t.Fatal("no nodes")
	}
	for _, node := range nodes {
		fmt.Println(node.GetXnet(), principal.Principal{Raw: node.NodeOperatorId})
	}
}

func TestClient_GetSubnetIDs(t *testing.T) {
	subnetIDs, err := client.GetSubnetIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(subnetIDs) == 0 {
		t.Fatal("no subnet IDs")
	}
	subnetID := subnetIDs[0]
	if _, err := client.GetSubnetDetails(subnetID); err != nil {
		t.Fatal(err)
	}
}
