package agent

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certificate"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
)

// ic0 is the old (default) host for the Internet Computer.
// var ic0, _ = url.Parse("https://ic0.app/")

// icp0 is the default host for the Internet Computer.
var icp0, _ = url.Parse("https://icp0.io/")

func uint64FromBytes(raw []byte) uint64 {
	switch len(raw) {
	case 1:
		return uint64(raw[0])
	case 2:
		return uint64(binary.BigEndian.Uint16(raw))
	case 4:
		return uint64(binary.BigEndian.Uint32(raw))
	case 8:
		return binary.BigEndian.Uint64(raw)
	default:
		panic(raw)
	}
}

// Agent is a client for the Internet Computer.
type Agent struct {
	client        Client
	identity      identity.Identity
	ingressExpiry time.Duration
	rootKey       []byte
}

// New returns a new Agent based on the given configuration.
func New(cfg Config) (*Agent, error) {
	if cfg.IngressExpiry == 0 {
		cfg.IngressExpiry = 10 * time.Second
	}
	// By default, use the anonymous identity.
	var id identity.Identity = new(identity.AnonymousIdentity)
	if cfg.Identity != nil {
		id = cfg.Identity
	}
	ccfg := ClientConfig{
		Host: icp0,
	}
	if cfg.ClientConfig != nil {
		ccfg = *cfg.ClientConfig
	}
	client := NewClient(ccfg)
	rootKey, _ := hex.DecodeString(certificate.RootKey)
	if cfg.FetchRootKey {
		status, err := client.Status()
		if err != nil {
			return nil, err
		}
		rootKey = status.RootKey
	}
	return &Agent{
		client:        client,
		identity:      id,
		ingressExpiry: cfg.IngressExpiry,
		rootKey:       rootKey,
	}, nil
}

// Call calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Call(canisterID principal.Principal, methodName string, args []any, values []any) error {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return err
	}
	if _, err := a.call(canisterID, data); err != nil {
		return err
	}
	raw, err := a.poll(canisterID, *requestID, time.Second, time.Second*10)
	if err != nil {
		return err
	}
	return idl.Unmarshal(raw, values)
}

// GetCanisterControllers returns the list of principals that can control the given canister.
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

// GetCanisterInfo returns the raw certificate for the given canister based on the given sub-path.
func (a Agent) GetCanisterInfo(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := [][]byte{[]byte("canister"), canisterID.Raw, []byte(subPath)}
	c, err := a.readStateCertificate(canisterID, [][][]byte{path})
	if err != nil {
		return nil, err
	}
	var state map[string]any
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, err
	}
	node, err := certificate.DeserializeNode(state["tree"].([]any))
	if err != nil {
		return nil, err
	}
	return certificate.Lookup(path, node), nil
}

// GetCanisterModuleHash returns the module hash for the given canister.
func (a Agent) GetCanisterModuleHash(canisterID principal.Principal) ([]byte, error) {
	return a.GetCanisterInfo(canisterID, "module_hash")
}

func (a Agent) Query(canisterID principal.Principal, methodName string, args []any, values []any) error {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return err
	}
	resp, err := a.query(canisterID, data)
	if err != nil {
		return err
	}
	var raw []byte
	switch resp.Status {
	case "replied":
		raw = resp.Reply["arg"]
	case "rejected":
		return fmt.Errorf("(%d) %s", resp.RejectCode, resp.RejectMsg)
	default:
		panic("unreachable")
	}
	return idl.Unmarshal(raw, values)
}

// RequestStatus returns the status of the request with the given ID.
func (a Agent) RequestStatus(canisterID principal.Principal, requestID RequestID) ([]byte, certificate.Node, error) {
	path := [][]byte{[]byte("request_status"), requestID[:]}
	c, err := a.readStateCertificate(canisterID, [][][]byte{path})
	if err != nil {
		return nil, nil, err
	}
	var state map[string]any
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, nil, err
	}
	cert, err := certificate.New(canisterID, a.rootKey[len(a.rootKey)-96:], c)
	if err != nil {
		return nil, nil, err
	}
	if err := cert.Verify(); err != nil {
		return nil, nil, err
	}
	node, err := certificate.DeserializeNode(state["tree"].([]any))
	if err != nil {
		return nil, nil, err
	}
	return certificate.Lookup(append(path, []byte("status")), node), node, nil
}

// Sender returns the principal that is sending the requests.
func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (a Agent) call(canisterID principal.Principal, data []byte) ([]byte, error) {
	return a.client.call(canisterID, data)
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) poll(canisterID principal.Principal, requestID RequestID, delay, timeout time.Duration) ([]byte, error) {
	ticker := time.NewTicker(delay)
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-ticker.C:
			data, node, err := a.RequestStatus(canisterID, requestID)
			if err != nil {
				return nil, err
			}
			if len(data) != 0 {
				path := [][]byte{[]byte("request_status"), requestID[:]}
				switch string(data) {
				case "rejected":
					code := certificate.Lookup(append(path, []byte("reject_code")), node)
					rejectMessage := certificate.Lookup(append(path, []byte("reject_message")), node)
					return nil, fmt.Errorf("(%d) %s", uint64FromBytes(code), string(rejectMessage))
				case "replied":
					path := [][]byte{[]byte("request_status"), requestID[:]}
					return certificate.Lookup(append(path, []byte("reply")), node), nil
				}
			}
		case <-timer.C:
			return nil, fmt.Errorf("out of time... waited %d seconds", timeout/time.Second)
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
		SenderPubKey: a.identity.PublicKey(),
		SenderSig:    requestID.Sign(a.identity),
	})
	if err != nil {
		return nil, nil, err
	}
	return &requestID, data, nil
}

// Config is the configuration for an Agent.
type Config struct {
	Identity      identity.Identity
	IngressExpiry time.Duration
	ClientConfig  *ClientConfig
	FetchRootKey  bool
}
