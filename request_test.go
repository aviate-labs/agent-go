package agent_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/principal-go"
)

func TestNewRequestID(t *testing.T) {
	// Source: https://smartcontracts.org/docs/interface-spec/index.html#request-id
	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Type:       agent.RequestTypeCall,
		CanisterID: principal.Principal{Raw: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xD2}},
		MethodName: "hello",
		Arguments:  []byte("DIDL\x00\xFD*"),
	})); h != "8781291c347db32a9d8c10eb62b710fce5a93be676474c42babc74c51858f94b" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Sender: principal.Principal{Raw: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xD2}},
		Paths: [][][]byte{
			{},
			{[]byte("")},
			{[]byte("hello"), []byte("world")},
		},
	})); h != "97d6f297aea699aec85d3377c7643ea66db810aba5c4372fbc2082c999f452dc" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Paths: [][][]byte{},
	})); h != "99daa8c80a61e87ac1fdf9dd49e39963bfe4dafb2a45095ebf4cad72d916d5be" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Paths: [][][]byte{{}},
	})); h != "ea01a9c3d3830db108e0a87995ea0d4183dc9c6e51324e9818fced5c57aa64f5" {
		t.Error(h)
	}
}
