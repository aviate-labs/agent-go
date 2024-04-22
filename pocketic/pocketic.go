package pocketic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
	"time"
)

var DefaultSubnetConfig = SubnetConfig{
	NNS: true,
}

type CanisterSettings struct {
	Controllers       *[]principal.Principal `ic:"controllers,omitempty" json:"controllers,omitempty"`
	ComputeAllocation *idl.Nat               `ic:"compute_allocation,omitempty" json:"compute_allocation,omitempty"`
	MemoryAllocation  *idl.Nat               `ic:"memory_allocation,omitempty" json:"memory_allocation,omitempty"`
	FreezingThreshold *idl.Nat               `ic:"freezing_threshold,omitempty" json:"freezing_threshold,omitempty"`
}

type CreateCanisterArgs struct {
	Settings    *CanisterSettings    `ic:"settings,omitempty" json:"settings,omitempty"`
	SpecifiedID *principal.Principal `ic:"specified_id" json:"specified_id,omitempty"`
}

type EffectiveCanisterID struct {
	CanisterId string `json:"CanisterId"`
}

type EffectiveSubnetID struct {
	SubnetID string `json:"SubnetId"`
}

type NNSConfig struct {
	StateDirPath string
	SubnetID     principal.Principal
}

type PocketIC struct {
	server     *server
	instanceID int
	topology   map[string]Topology
	sender     principal.Principal
}

func (pic PocketIC) UpdateCall(canisterID principal.Principal, method string, payload []any, body []any) error {
	rawPayload, err := idl.Marshal(payload)
	if err != nil {
		return err
	}
	return pic.UpdateCallWithEffectiveCanisterID(&canisterID, nil, method, rawPayload, body)
}

func (pic PocketIC) QueryCall(canisterID principal.Principal, method string, payload []any, body []any) error {
	rawPayload, err := idl.Marshal(payload)
	if err != nil {
		return err
	}
	return pic.canisterCall("read/query", &canisterID, nil, method, rawPayload, body)
}

func (pic PocketIC) CreateAndInstallCanister(wasmModule []byte, arg []byte, subnetPID *principal.Principal) (*principal.Principal, error) {
	canisterID, err := pic.CreateCanister(CreateCanisterArgs{}, subnetPID)
	if err != nil {
		return nil, err
	}
	if _, err := pic.AddCycles(*canisterID, 2_000_000_000_000); err != nil {
		return nil, err
	}
	if err := pic.InstallCode(*canisterID, wasmModule, arg); err != nil {
		return nil, err
	}
	return canisterID, nil
}

// New creates a new PocketIC instance with the given subnet configuration.
func New(subnetConfig SubnetConfig) (*PocketIC, error) {
	s, err := newServer()
	if err != nil {
		return nil, err
	}
	if !subnetConfig.validate() {
		return nil, fmt.Errorf("invalid subnet config")
	}
	resp, err := s.NewInstance(subnetConfig)
	if err != nil {
		return nil, err
	}
	return &PocketIC{
		server:     s,
		instanceID: resp.InstanceID,
		topology:   resp.Topology,
		sender:     principal.AnonymousID,
	}, nil
}

func (pic PocketIC) AddCycles(canisterID principal.Principal, amount int) (int, error) {
	var body struct {
		Cycles int `json:"cycles"`
	}
	if err := pic.server.InstancePost(pic.instanceID, "update/add_cycles", map[string]any{
		"canister_id": base64.StdEncoding.EncodeToString(canisterID.Raw),
		"amount":      amount,
	}, &body); err != nil {
		return 0, err
	}
	return body.Cycles, nil
}

// AdvanceTime advances the time of the PocketIC instance by the given nanoseconds.
func (pic PocketIC) AdvanceTime(nanoSeconds int) error {
	t, err := pic.GetTime()
	if err != nil {
		return err
	}
	return pic.server.InstancePost(pic.instanceID, "update/set_time", map[string]any{
		"nanos_since_epoch": t.Nanosecond() + nanoSeconds,
	}, nil)
}

// CanisterExits returns true if the given canister exists in the PocketIC instance.
func (pic PocketIC) CanisterExits(canisterID principal.Principal) bool {
	_, err := pic.GetSubnet(canisterID)
	return err == nil
}

func (pic PocketIC) CreateCanister(args CreateCanisterArgs, subnetPID *principal.Principal) (*principal.Principal, error) {
	var ecID any
	if subnetPID != nil {
		ecID = EffectiveSubnetID{
			SubnetID: base64.StdEncoding.EncodeToString(subnetPID.Raw),
		}
	}

	payload, err := idl.Marshal([]any{args})
	if err != nil {
		return nil, err
	}

	var resp struct {
		CanisterID principal.Principal `ic:"canister_id"`
	}
	if err := pic.UpdateCallWithEffectiveCanisterID(
		nil,
		ecID,
		"provisional_create_canister_with_cycles",
		payload,
		[]any{&resp},
	); err != nil {
		return nil, err
	}
	return &resp.CanisterID, nil
}

func (pic PocketIC) GetCycleBalance(canisterID principal.Principal) (int, error) {
	var body struct {
		Cycles int `json:"cycles"`
	}
	if err := pic.server.InstancePost(pic.instanceID, "read/get_cycles", map[string]any{
		"canister_id": base64.StdEncoding.EncodeToString(canisterID.Raw),
	}, &body); err != nil {
		return 0, err
	}
	return body.Cycles, nil
}

// GetRootKey returns the root key of the PocketIC instance.
func (pic PocketIC) GetRootKey() ([]byte, error) {
	var nnsPID principal.Principal
	for k, v := range pic.topology {
		if v.SubnetKind == NNSSubnet {
			pid, err := principal.Decode(k)
			if err != nil {
				return nil, err
			}
			nnsPID = pid
			break
		}

	}
	if len(nnsPID.Raw) == 0 {
		return nil, fmt.Errorf("NNS subnet not found")
	}
	var body []byte
	if err := pic.server.InstancePost(pic.instanceID, "read/pub_key", map[string]any{
		"subnet_id": base64.StdEncoding.EncodeToString(nnsPID.Raw),
	}, &body); err != nil {
		return nil, err
	}
	return body, nil
}

// GetSubnet returns the subnet of the given canister.
func (pic PocketIC) GetSubnet(canisterID principal.Principal) (*principal.Principal, error) {
	var body struct {
		SubnetID string `json:"subnet_id"`
	}
	if err := pic.server.InstancePost(pic.instanceID, "read/get_subnet", map[string]any{
		"canister_id": base64.StdEncoding.EncodeToString(canisterID.Raw),
	}, &body); err != nil {
		return nil, err
	}
	subnetPID, err := base64.StdEncoding.DecodeString(body.SubnetID)
	if err != nil {
		return nil, err
	}
	return &principal.Principal{
		Raw: subnetPID,
	}, nil
}

// GetTime returns the current time of the PocketIC instance.
func (pic PocketIC) GetTime() (*time.Time, error) {
	var m struct {
		NanosSinceEpoch int64 `json:"nanos_since_epoch"`
	}
	if err := pic.server.InstanceGet(pic.instanceID, "read/get_time", &m); err != nil {
		return nil, err
	}
	t := time.Unix(0, m.NanosSinceEpoch)
	return &t, nil
}

func (pic PocketIC) InstallCode(canisterID principal.Principal, wasmModule []byte, arg []byte) error {
	payload, err := idl.Marshal([]any{installCodeArgs{
		WasmModule: wasmModule,
		CanisterID: canisterID,
		Arg:        arg,
		Mode: installMode{
			Install: &idl.Null{},
		},
	}})
	if err != nil {
		return err
	}
	return pic.UpdateCallWithEffectiveCanisterID(
		nil,
		EffectiveCanisterID{
			CanisterId: base64.StdEncoding.EncodeToString(canisterID.Raw),
		},
		"install_code",
		payload,
		nil,
	)
}

// SetSender sets the sender principal for the PocketIC instance.
func (pic *PocketIC) SetSender(sender principal.Principal) {
	pic.sender = sender
}

// SetTime sets the time of the PocketIC instance to the given nanoseconds since epoch.
func (pic PocketIC) SetTime(nanosSinceEpoch int) error {
	return pic.server.InstancePost(pic.instanceID, "update/set_time", map[string]any{
		"nanos_since_epoch": nanosSinceEpoch,
	}, nil)
}

// Tick advances the PocketIC instance by one block.
func (pic PocketIC) Tick() error {
	return pic.server.InstancePost(pic.instanceID, "update/tick", nil, nil)
}

func (pic PocketIC) UpdateCallWithEffectiveCanisterID(canisterID *principal.Principal, ecID any, method string, payload []byte, body []any) error {
	return pic.canisterCall("update/execute_ingress_message", canisterID, ecID, method, payload, body)
}

func (pic PocketIC) canisterCall(endpoint string, canisterID *principal.Principal, ecID any, method string, payload []byte, body []any) error {
	if ecID == nil {
		ecID = "None"
	}
	var cID principal.Principal
	if canisterID != nil {
		cID = *canisterID
	}
	var reply reply
	if err := pic.server.InstancePost(pic.instanceID, endpoint, map[string]any{
		"sender":              base64.StdEncoding.EncodeToString(pic.sender.Raw),
		"effective_principal": ecID,
		"canister_id":         base64.StdEncoding.EncodeToString(cID.Raw),
		"method":              method,
		"payload":             base64.StdEncoding.EncodeToString(payload),
	}, &reply); err != nil {
		return err
	}
	if reply.Ok == nil {
		return reply.Err
	}
	if reply.Ok.Reply == nil {
		return *reply.Ok.Reject
	}
	rawBody, err := base64.StdEncoding.DecodeString(*reply.Ok.Reply)
	if err != nil {
		return err
	}
	return idl.Unmarshal(rawBody, body)
}

type RejectError string

func (e RejectError) Error() string {
	return string(e)
}

type ReplyError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (e ReplyError) Error() string {
	return fmt.Sprintf("code: %d, description: %s", e.Code, e.Description)
}

type SubnetConfig struct {
	Application uint
	Bitcoin     bool
	Fiduciary   bool
	II          bool
	NNS         bool
	NNSConfig   *NNSConfig
	SNS         bool
	System      uint
}

func (s SubnetConfig) MarshalJSON() ([]byte, error) {
	newBool := func(b bool) *string {
		if b {
			n := "New"
			return &n
		}
		return nil
	}
	newUint := func(u uint) []string {
		n := make([]string, 0, u)
		for i := uint(0); i < u; i++ {
			n = append(n, "New")
		}
		return n
	}
	newNNS := func(b bool, config *NNSConfig) any {
		if config != nil {
			return map[string]interface{}{
				"FromPath":  config.StateDirPath,
				"subnet-id": config.SubnetID,
			}
		}
		return newBool(b)
	}
	return json.Marshal(map[string]interface{}{
		"application": newUint(s.Application),
		"bitcoin":     newBool(s.Bitcoin),
		"fiduciary":   newBool(s.Fiduciary),
		"ii":          newBool(s.II),
		"nns":         newNNS(s.NNS, s.NNSConfig),
		"sns":         newBool(s.SNS),
		"system":      newUint(s.System),
	})
}

func (s SubnetConfig) validate() bool {
	// At least one subnet must be enabled.
	return 0 < s.Application || s.Bitcoin || s.Fiduciary || s.II || s.NNS || s.SNS || 0 < s.System
}

type SubnetKind string

var (
	ApplicationSubnet SubnetKind = "Application"
	BitcoinSubnet     SubnetKind = "Bitcoin"
	FiduciarySubnet   SubnetKind = "Fiduciary"
	IISubnet          SubnetKind = "II"
	NNSSubnet         SubnetKind = "NNS"
	SNSSubnet         SubnetKind = "SNS"
	SystemSubnet      SubnetKind = "System"
)

type installCodeArgs struct {
	WasmModule []byte              `ic:"wasm_module"`
	CanisterID principal.Principal `ic:"canister_id"`
	Arg        []byte              `ic:"arg"`
	Mode       installMode         `ic:"mode"`
}

type installMode struct {
	Install   *idl.Null `ic:"install,variant"`
	Reinstall *idl.Null `ic:"reinstall,variant"`
	Upgrade   *idl.Null `ic:"upgrade,variant"`
}

type reply struct {
	Ok *struct {
		Reply  *string      `json:"Reply,omitempty"`
		Reject *RejectError `json:"Reject,omitempty"`
	} `json:"Ok,omitempty"`
	Err *ReplyError `json:"Err,omitempty"`
}
