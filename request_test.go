package agent_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go"
)

func TestNewRequestID(t *testing.T) {
	req := agent.Request{
		Type:       agent.RequestTypeCall,
		CanisterID: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xD2},
		MethodName: "hello",
		Arguments:  []byte("DIDL\x00\xFD*"),
	}
	h := fmt.Sprintf("%x", agent.NewRequestID(req))
	if h != "8781291c347db32a9d8c10eb62b710fce5a93be676474c42babc74c51858f94b" {
		t.Fatal(h)
	}
}
