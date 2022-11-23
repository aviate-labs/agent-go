package agent

import (
	"encoding/binary"
	"fmt"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go/candid"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certificate"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

var ic0, _ = url.Parse("https://ic0.app/")

type Agent struct {
	client        Client
	identity      identity.Identity
	ingressExpiry time.Duration
}

func New(cfg AgentConfig) Agent {
	if cfg.IngressExpiry == 0 {
		cfg.IngressExpiry = 10 * time.Second
	}
	var id identity.Identity = identity.AnonymousIdentity{}
	if cfg.Identity != nil {
		id = *cfg.Identity
	}
	ccfg := ClientConfig{
		Host: ic0,
	}
	if cfg.ClientConfig != nil {
		ccfg = *cfg.ClientConfig
	}
	return Agent{
		client:        NewClient(ccfg),
		identity:      id,
		ingressExpiry: cfg.IngressExpiry,
	}
}

func (a Agent) Call(canisterID principal.Principal, methodName string, args []byte) (string, error) {
	types, values, err := a.CallCandid(canisterID, methodName, args)
	if err != nil {
		return "", err
	}
	return candid.DecodeValues(types, values)
}

func (a Agent) CallCandid(canisterID principal.Principal, methodName string, args []byte) ([]idl.Type, []interface{}, error) {
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     args,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, nil, err
	}
	if _, err := a.call(canisterID, data); err != nil {
		return nil, nil, err
	}
	return a.poll(canisterID, *requestID, time.Second, time.Second*10)
}

func (a Agent) GetCanisterControllers(canisterID principal.Principal) ([]principal.Principal, error) {
	resp, err := a.GetCanisterInfo(canisterID, "controllers")
	if err != nil {
		return nil, err
	}
	var m [][]byte
	if err := cbor.Unmarshal(resp, &m); err != nil {
		return nil, err
	}
	var p []principal.Principal
	for _, b := range m {
		p = append(p, principal.Principal{Raw: b})
	}
	return p, nil
}

func (a Agent) GetCanisterInfo(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := [][]byte{[]byte("canister"), canisterID.Raw, []byte(subPath)}
	c, err := a.readStateCertificate(canisterID, [][][]byte{path})
	if err != nil {
		return nil, err
	}
	var state map[string]interface{}
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, err
	}
	node, err := certificate.DeserializeNode(state["tree"].([]interface{}))
	if err != nil {
		return nil, err
	}
	return certificate.Lookup(path, node), nil
}

func (a Agent) GetCanisterModuleHash(canisterID principal.Principal) ([]byte, error) {
	return a.GetCanisterInfo(canisterID, "module_hash")
}

func (a Agent) Query(canisterID principal.Principal, methodName string, args []byte) (string, error) {
	types, values, err := a.QueryCandid(canisterID, methodName, args)
	if err != nil {
		return "", err
	}
	return candid.DecodeValues(types, values)
}

func (a Agent) QueryCandid(canisterID principal.Principal, methodName string, args []byte) ([]idl.Type, []interface{}, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     args,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, nil, err
	}
	resp, err := a.query(canisterID, data)
	if err != nil {
		return nil, nil, err
	}
	switch resp.Status {
	case "replied":
		return idl.Decode(resp.Reply["arg"])
	case "rejected":
		return nil, nil, fmt.Errorf("(%d) %s", resp.RejectCode, resp.RejectMsg)
	default:
		panic("unreachable")
	}
}

func (agent *Agent) RequestStatus(canisterID principal.Principal, requestID RequestID) ([]byte, certificate.Node, error) {
	path := [][]byte{[]byte("request_status"), requestID[:]}
	c, err := agent.readStateCertificate(canisterID, [][][]byte{path})
	if err != nil {
		return nil, nil, err
	}
	var state map[string]interface{}
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, nil, err
	}
	node, err := certificate.DeserializeNode(state["tree"].([]interface{}))
	if err != nil {
		return nil, nil, err
	}
	return certificate.Lookup(append(path, []byte("status")), node), node, nil
}

func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (agent *Agent) call(canisterID principal.Principal, data []byte) ([]byte, error) {
	return agent.client.call(canisterID, data)
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) poll(canisterID principal.Principal, requestID RequestID, delay, timeout time.Duration) ([]idl.Type, []interface{}, error) {
	ticker := time.NewTicker(delay)
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-ticker.C:
			data, node, err := a.RequestStatus(canisterID, requestID)
			if err != nil {
				return nil, nil, err
			}
			if len(data) != 0 {
				path := [][]byte{[]byte("request_status"), requestID[:]}
				switch string(data) {
				case "rejected":
					code := certificate.Lookup(append(path, []byte("reject_code")), node)
					reject_message := certificate.Lookup(append(path, []byte("reject_message")), node)
					return nil, nil, fmt.Errorf("(%d) %s", binary.BigEndian.Uint64(code), string(reject_message))
				case "replied":
					path := [][]byte{[]byte("request_status"), requestID[:]}
					reply := certificate.Lookup(append(path, []byte("reply")), node)
					return idl.Decode(reply)
				}
			}
		case <-timer.C:
			return nil, nil, fmt.Errorf("out of time... waited %d seconds", timeout/time.Second)
		}
	}
}

func (a Agent) query(canisterID principal.Principal, data []byte) (*Response, error) {
	resp, err := a.client.query(canisterID, data)
	if err != nil {
		return nil, err
	}
	queryResponse := new(Response)
	return queryResponse, cbor.Unmarshal(resp, queryResponse)
}

func (a Agent) readState(canisterID principal.Principal, data []byte) (map[string][]byte, error) {
	resp, err := a.client.readState(canisterID, data)
	if err != nil {
		return nil, err
	}
	var m map[string][]byte
	return m, cbor.Unmarshal(resp, &m)
}

func (a Agent) readStateCertificate(canisterID principal.Principal, paths [][][]byte) ([]byte, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeReadState,
		Sender:        a.Sender(),
		Paths:         paths,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, err
	}

	resp, err := a.readState(canisterID, data)
	if err != nil {
		return nil, err
	}
	return resp["certificate"], nil
}

func (a Agent) sign(request Request) (*RequestID, []byte, error) {
	requestID := NewRequestID(request)
	data, err := cbor.Marshal(Envelope{
		Content:      request,
		SenderPubkey: a.identity.PublicKey(),
		SenderSig:    requestID.Sign(a.identity),
	})
	if err != nil {
		return nil, nil, err
	}
	return &requestID, data, nil
}

type AgentConfig struct {
	Identity      *identity.Identity
	IngressExpiry time.Duration
	ClientConfig  *ClientConfig
}
