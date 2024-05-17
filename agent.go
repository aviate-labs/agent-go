package agent

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net/url"
	"reflect"
	"time"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/certification"
	"github.com/aviate-labs/agent-go/certification/hashtree"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/agent-go/principal"

	"github.com/fxamacker/cbor/v2"
)

// DefaultConfig is the default configuration for an Agent.
var DefaultConfig = Config{}

// ic0 is the old (default) host for the Internet Computer.
// var ic0, _ = url.Parse("https://ic0.app/")

// icp0 is the default host for the Internet Computer.
var icp0, _ = url.Parse("https://icp0.io/")

func effectiveCanisterID(canisterID principal.Principal, args []any) principal.Principal {
	// If the canisterID is not aaaaa-aa (encoded as empty byte array), return it.
	if 0 < len(canisterID.Raw) || len(args) < 1 {
		return canisterID
	}

	v := reflect.ValueOf(args[0])
	switch v.Kind() {
	case reflect.Map:
		if ecid, ok := args[0].(map[string]any)["canister_id"]; ok {
			switch ecidp := ecid.(type) {
			case principal.Principal:
				return ecidp
			default:
				// If the field is not a principal, return the original canisterId.
				return canisterID
			}
		}
		return canisterID
	case reflect.Struct:
		t := v.Type()
		// Get the field with the ic tag "canister_id".
		for idx := range t.NumField() {
			if tag := t.Field(idx).Tag.Get("ic"); tag == "canister_id" {
				ecid := v.Field(idx).Interface()
				switch ecidp := ecid.(type) {
				case principal.Principal:
					return ecidp
				default:
					// If the field is not a principal, return the original canisterId.
					return canisterID
				}
			}
		}
		return canisterID
	default:
		return canisterID
	}
}

func newNonce() ([]byte, error) {
	/* Read 10 bytes of random data, which is smaller than the max allowed by the IC (32 bytes)
	 * and should still be enough from a practical point of view. */
	nonce := make([]byte, 10)
	_, err := rand.Read(nonce)
	return nonce, err
}

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
	client         Client
	identity       identity.Identity
	ingressExpiry  time.Duration
	rootKey        []byte
	logger         Logger
	delay, timeout time.Duration
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
	var logger Logger = new(NoopLogger)
	if cfg.Logger != nil {
		logger = cfg.Logger
	}
	ccfg := ClientConfig{
		Host: icp0,
	}
	if cfg.ClientConfig != nil {
		ccfg = *cfg.ClientConfig
	}
	client := NewClientWithLogger(ccfg, logger)
	rootKey, _ := hex.DecodeString(certification.RootKey)
	if cfg.FetchRootKey {
		status, err := client.Status()
		if err != nil {
			return nil, err
		}
		rootKey = status.RootKey
	}
	delay := time.Second
	if cfg.PollDelay != 0 {
		delay = cfg.PollDelay
	}
	timeout := 10 * time.Second
	if cfg.PollTimeout != 0 {
		timeout = cfg.PollTimeout
	}
	return &Agent{
		client:        client,
		identity:      id,
		ingressExpiry: cfg.IngressExpiry,
		rootKey:       rootKey,
		logger:        logger,
		delay:         delay,
		timeout:       timeout,
	}, nil
}

// Call calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Call(canisterID principal.Principal, methodName string, args []any, values []any) error {
	call, err := a.CreateCall(canisterID, methodName, args...)
	if err != nil {
		return err
	}
	return call.CallAndWait(values...)
}

// CallProto calls a method on a canister and unmarshals the result into the given proto message.
func (a Agent) CallProto(canisterID principal.Principal, methodName string, in, out proto.Message) error {
	payload, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		IngressExpiry: a.expiryDate(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     payload,
	})
	if err != nil {
		return err
	}
	if _, err := a.call(canisterID, data); err != nil {
		return err
	}
	raw, err := a.poll(canisterID, *requestID)
	if err != nil {
		return err
	}
	return proto.Unmarshal(raw, out)
}

// Client returns the underlying Client of the Agent.
func (a Agent) Client() *Client {
	return &a.client
}

// CreateCall creates a new Call to the given canister and method.
func (a *Agent) CreateCall(canisterID principal.Principal, methodName string, args ...any) (*Call, error) {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	requestID, data, err := a.sign(Request{
		Type:          RequestTypeCall,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
		Nonce:         nonce,
	})
	if err != nil {
		return nil, err
	}
	return &Call{
		a:                   a,
		methodName:          methodName,
		effectiveCanisterID: effectiveCanisterID(canisterID, args),
		requestID:           *requestID,
		data:                data,
	}, nil
}

// CreateQuery creates a new Query to the given canister and method.
func (a *Agent) CreateQuery(canisterID principal.Principal, methodName string, args ...any) (*Query, error) {
	rawArgs, err := idl.Marshal(args)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		// Default to the empty Candid argument list.
		rawArgs = []byte{'D', 'I', 'D', 'L', 0, 0}
	}
	nonce, err := newNonce()
	if err != nil {
		return nil, err
	}
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     rawArgs,
		IngressExpiry: a.expiryDate(),
		Nonce:         nonce,
	})
	if err != nil {
		return nil, err
	}
	return &Query{
		a:                   a,
		methodName:          methodName,
		effectiveCanisterID: effectiveCanisterID(canisterID, args),
		data:                data,
	}, nil
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
	path := []hashtree.Label{hashtree.Label("canister"), canisterID.Raw, hashtree.Label(subPath)}
	node, err := a.ReadStateCertificate(canisterID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	canisterInfo, err := hashtree.NewHashTree(node).Lookup(path...)
	if err != nil {
		return nil, err
	}
	return canisterInfo, nil
}

func (a Agent) GetCanisterMetadata(canisterID principal.Principal, subPath string) ([]byte, error) {
	path := []hashtree.Label{hashtree.Label("canister"), canisterID.Raw, hashtree.Label("metadata"), hashtree.Label(subPath)}
	c, err := a.readStateCertificate(canisterID, [][]hashtree.Label{path})
	if err != nil {
		return nil, err
	}
	var state map[string]any
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, err
	}
	node, err := hashtree.DeserializeNode(state["tree"].([]any))
	if err != nil {
		return nil, err
	}
	metadata, err := hashtree.NewHashTree(node).Lookup(path...)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// GetCanisterModuleHash returns the module hash for the given canister.
func (a Agent) GetCanisterModuleHash(canisterID principal.Principal) ([]byte, error) {
	h, err := a.GetCanisterInfo(canisterID, "module_hash")
	var lookupError hashtree.LookupError
	if errors.As(err, &lookupError) && lookupError.Type == hashtree.LookupResultAbsent {
		// If the canister is empty, it is expected that the module hash is not available.
		return nil, nil
	}
	return h, err
}

// GetRootKey returns the root key of the host.
func (a Agent) GetRootKey() []byte {
	return a.rootKey
}

// Query calls a method on a canister and unmarshals the result into the given values.
func (a Agent) Query(canisterID principal.Principal, methodName string, args, values []any) error {
	query, err := a.CreateQuery(canisterID, methodName, args...)
	if err != nil {
		return err
	}
	return query.Query(values...)
}

// QueryProto calls a method on a canister and unmarshals the result into the given proto message.
func (a Agent) QueryProto(canisterID principal.Principal, methodName string, in, out proto.Message) error {
	payload, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	_, data, err := a.sign(Request{
		Type:          RequestTypeQuery,
		Sender:        a.Sender(),
		IngressExpiry: a.expiryDate(),
		CanisterID:    canisterID,
		MethodName:    methodName,
		Arguments:     payload,
	})
	if err != nil {
		return err
	}
	resp, err := a.client.Query(canisterID, data)
	if err != nil {
		return err
	}
	var response Response
	if err := cbor.Unmarshal(resp, &response); err != nil {
		return err
	}
	if response.Status != "replied" {
		return fmt.Errorf("status: %s", response.Status)
	}
	return proto.Unmarshal(response.Reply["arg"], out)
}

// ReadStateCertificate reads the certificate state of the given canister at the given path.
func (a Agent) ReadStateCertificate(canisterID principal.Principal, path [][]hashtree.Label) (hashtree.Node, error) {
	c, err := a.readStateCertificate(canisterID, path)
	if err != nil {
		return nil, err
	}
	var state map[string]any
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, err
	}
	return hashtree.DeserializeNode(state["tree"].([]any))
}

// RequestStatus returns the status of the request with the given ID.
func (a Agent) RequestStatus(ecID principal.Principal, requestID RequestID) ([]byte, hashtree.Node, error) {
	a.logger.Printf("[AGENT] REQUEST STATUS %s %x", ecID, requestID)
	path := []hashtree.Label{hashtree.Label("request_status"), requestID[:]}
	c, err := a.readStateCertificate(ecID, [][]hashtree.Label{path})
	if err != nil {
		return nil, nil, err
	}
	var state map[string]any
	if err := cbor.Unmarshal(c, &state); err != nil {
		return nil, nil, err
	}
	cert, err := certification.New(ecID, a.rootKey[len(a.rootKey)-96:], c)
	if err != nil {
		return nil, nil, err
	}
	if err := cert.Verify(); err != nil {
		return nil, nil, err
	}
	node, err := hashtree.DeserializeNode(state["tree"].([]any))
	if err != nil {
		return nil, nil, err
	}
	status, err := hashtree.NewHashTree(node).Lookup(append(path, hashtree.Label("status"))...)
	var lookupError hashtree.LookupError
	if errors.As(err, &lookupError) && lookupError.Type == hashtree.LookupResultAbsent {
		// The status might not be available immediately, since the request is still being processed.
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return status, node, nil
}

// Sender returns the principal that is sending the requests.
func (a Agent) Sender() principal.Principal {
	return a.identity.Sender()
}

func (a Agent) call(ecID principal.Principal, data []byte) ([]byte, error) {
	return a.client.Call(ecID, data)
}

func (a Agent) expiryDate() uint64 {
	return uint64(time.Now().Add(a.ingressExpiry).UnixNano())
}

func (a Agent) poll(ecID principal.Principal, requestID RequestID) ([]byte, error) {
	ticker := time.NewTicker(a.delay)
	timer := time.NewTimer(a.timeout)
	for {
		select {
		case <-ticker.C:
			a.logger.Printf("[AGENT] POLL %s %x", ecID, requestID)
			data, node, err := a.RequestStatus(ecID, requestID)
			if err != nil {
				return nil, err
			}
			if len(data) != 0 {
				path := []hashtree.Label{hashtree.Label("request_status"), requestID[:]}
				switch string(data) {
				case "rejected":
					tree := hashtree.NewHashTree(node)
					code, err := tree.Lookup(append(path, hashtree.Label("reject_code"))...)
					if err != nil {
						return nil, err
					}
					message, err := tree.Lookup(append(path, hashtree.Label("reject_message"))...)
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf("(%d) %s", uint64FromBytes(code), string(message))
				case "replied":
					replied, err := hashtree.NewHashTree(node).Lookup(append(path, hashtree.Label("reply"))...)
					if err != nil {
						return nil, fmt.Errorf("no reply found")
					}
					return replied, nil
				}
			}
		case <-timer.C:
			return nil, fmt.Errorf("out of time... waited %d seconds", a.timeout/time.Second)
		}
	}
}

func (a Agent) query(canisterID principal.Principal, data []byte) (*Response, error) {
	resp, err := a.client.Query(canisterID, data)
	if err != nil {
		return nil, err
	}
	queryResponse := new(Response)
	return queryResponse, cbor.Unmarshal(resp, queryResponse)
}

func (a Agent) readState(ecID principal.Principal, data []byte) (map[string][]byte, error) {
	resp, err := a.client.ReadState(ecID, data)
	if err != nil {
		return nil, err
	}
	var m map[string][]byte
	return m, cbor.Unmarshal(resp, &m)
}

func (a Agent) readStateCertificate(ecID principal.Principal, paths [][]hashtree.Label) ([]byte, error) {
	_, data, err := a.sign(Request{
		Type:          RequestTypeReadState,
		Sender:        a.Sender(),
		Paths:         paths,
		IngressExpiry: a.expiryDate(),
	})
	if err != nil {
		return nil, err
	}
	a.logger.Printf("[AGENT] READ STATE %s (ecID)", ecID)
	resp, err := a.readState(ecID, data)
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

// Call is an intermediate representation of a Call to a canister.
type Call struct {
	a                   *Agent
	methodName          string
	effectiveCanisterID principal.Principal
	requestID           RequestID
	data                []byte
}

// Call calls a method on a canister, it does not wait for the result.
func (c Call) Call() error {
	c.a.logger.Printf("[AGENT] CALL %s %s (%x)", c.effectiveCanisterID, c.methodName, c.requestID)
	_, err := c.a.call(c.effectiveCanisterID, c.data)
	return err
}

// CallAndWait calls a method on a canister and waits for the result.
func (c Call) CallAndWait(values ...any) error {
	if err := c.Call(); err != nil {
		return err
	}
	return c.Wait(values...)
}

// Wait waits for the result of the Call and unmarshals it into the given values.
func (c Call) Wait(values ...any) error {
	raw, err := c.a.poll(c.effectiveCanisterID, c.requestID)
	if err != nil {
		return err
	}
	return idl.Unmarshal(raw, values)
}

// WithEffectiveCanisterID sets the effective canister ID for the Call.
func (c *Call) WithEffectiveCanisterID(canisterID principal.Principal) *Call {
	c.effectiveCanisterID = canisterID
	return c
}

// Config is the configuration for an Agent.
type Config struct {
	// Identity is the identity used by the Agent.
	Identity identity.Identity
	// IngressExpiry is the duration for which an ingress message is valid.
	IngressExpiry time.Duration
	// ClientConfig is the configuration for the underlying Client.
	ClientConfig *ClientConfig
	// FetchRootKey determines whether the root key should be fetched from the IC.
	FetchRootKey bool
	// Logger is the logger used by the Agent.
	Logger Logger
	// PollDelay is the delay between polling for a response.
	PollDelay time.Duration
	// PollTimeout is the timeout for polling for a response.
	PollTimeout time.Duration
}

// Query is an intermediate representation of a Query to a canister.
type Query struct {
	a                   *Agent
	methodName          string
	effectiveCanisterID principal.Principal
	data                []byte
}

// Query calls a method on a canister and unmarshals the result into the given values.
func (q Query) Query(values ...any) error {
	q.a.logger.Printf("[AGENT] QUERY %s %s", q.effectiveCanisterID, q.methodName)
	resp, err := q.a.query(q.effectiveCanisterID, q.data)
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
