package agent_test

import (
	"fmt"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/principal"
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
		Paths: [][]hashtree.Label{
			{},
			{[]byte("")},
			{[]byte("hello"), []byte("world")},
		},
	})); h != "97d6f297aea699aec85d3377c7643ea66db810aba5c4372fbc2082c999f452dc" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Paths: [][]hashtree.Label{},
	})); h != "99daa8c80a61e87ac1fdf9dd49e39963bfe4dafb2a45095ebf4cad72d916d5be" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Paths: [][]hashtree.Label{{}},
	})); h != "ea01a9c3d3830db108e0a87995ea0d4183dc9c6e51324e9818fced5c57aa64f5" {
		t.Error(h)
	}

	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Type:          agent.RequestTypeCall,
		Sender:        principal.AnonymousID,
		IngressExpiry: 1711532558242940000,
		CanisterID:    principal.Principal{Raw: make([]byte, 0)}, // aaaaa-aa
		MethodName:    "update_settings",
		Arguments: []byte{
			// ic0.UpdateSettingsArgs{
			// 	CanisterId: "bkyz2-fmaaa-aaaaa-qaaaq-cai",
			//	Settings: ic0.CanisterSettings{
			//		Controllers: &[]principal.Principal{
			//			principal.AnonymousID,
			//		},
			//	},
			// }
			0x44, 0x49, 0x44, 0x4c, 0x06, 0x6e, 0x7d, 0x6d, 0x68, 0x6e, 0x01, 0x6c, 0x05, 0xc0, 0xcf, 0xf2,
			0x71, 0x00, 0xd7, 0xe0, 0x9b, 0x90, 0x02, 0x02, 0x80, 0xad, 0x98, 0x8a, 0x04, 0x00, 0xde, 0xeb,
			0xb5, 0xa9, 0x0e, 0x00, 0xa8, 0x82, 0xac, 0xc6, 0x0f, 0x00, 0x6e, 0x78, 0x6c, 0x03, 0xb3, 0xc4,
			0xb1, 0xf2, 0x04, 0x68, 0xe3, 0xf9, 0xf5, 0xd9, 0x08, 0x03, 0xca, 0x99, 0x98, 0xb4, 0x0d, 0x04,
			0x01, 0x05, 0x01, 0x0a, 0x80, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x01, 0x01, 0x01, 0x00, 0x01,
			0x01, 0x01, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00,
		},
	})); h != "3599fd3f4505a6ec44429dddff35a3e1338d9d28c64444cf4632df427d83d3cf" {
		t.Error(h)
	}
}

func TestRequestID_Sign(t *testing.T) {
	if h := fmt.Sprintf("%x", agent.NewRequestID(agent.Request{
		Type:          agent.RequestTypeCall,
		Sender:        principal.AnonymousID,
		CanisterID:    principal.Principal{Raw: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xD2}},
		MethodName:    "hello",
		IngressExpiry: 1685570400000000000,
		Arguments:     []byte("DIDL\x00\xFD*"),
	})); h != "1d1091364d6bb8a6c16b203ee75467d59ead468f523eb058880ae8ec80e2b101" {
		t.Error(h)
	}
}
