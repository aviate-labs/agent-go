package pocketic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/principal"
)

var (
	DefaultSubnetSpec = SubnetSpec{
		StateConfig:       SubnetStateConfigNew{},
		InstructionConfig: SubnetInstructionConfigProduction{},
		DTSFlag:           false,
	}
	DefaultSubnetConfig = SubnetConfigSet{
		NNS: &DefaultSubnetSpec,
		// The JSON API requires an empty array for the Application and System subnets.
		Application: make([]SubnetSpec, 0),
		System:      make([]SubnetSpec, 0),
	}
)

type CanisterIDRange struct {
	Start principal.Principal
	End   principal.Principal
}

func (c CanisterIDRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(canisterIDRange{
		Start: rawCanisterID{CanisterID: base64.StdEncoding.EncodeToString(c.Start.Raw)},
		End:   rawCanisterID{CanisterID: base64.StdEncoding.EncodeToString(c.End.Raw)},
	})
}

func (c *CanisterIDRange) UnmarshalJSON(bytes []byte) error {
	var r canisterIDRange
	if err := json.Unmarshal(bytes, &r); err != nil {
		return err
	}
	start, err := base64.StdEncoding.DecodeString(r.Start.CanisterID)
	if err != nil {
		return err
	}
	c.Start = principal.Principal{Raw: start}
	end, err := base64.StdEncoding.DecodeString(r.End.CanisterID)
	if err != nil {
		return err
	}
	c.End = principal.Principal{Raw: end}
	return nil
}

type Config struct {
	subnetConfig   SubnetConfigSet
	serverConfig   []serverOption
	client         *http.Client
	logger         agent.Logger
	delay, timeout time.Duration
}

type DTSFlag bool

func (f DTSFlag) MarshalJSON() ([]byte, error) {
	if f {
		return json.Marshal("Enabled")
	}
	return json.Marshal("Disabled")
}

func (f *DTSFlag) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	if s != "Enabled" && s != "Disabled" {
		return fmt.Errorf("invalid DTS flag: %s", s)
	}
	*f = s == "Enabled"
	return nil
}

type Option func(*Config)

// WithApplicationSubnet adds an empty Application subnet.
func WithApplicationSubnet() Option {
	return func(p *Config) {
		p.subnetConfig.Application = append(p.subnetConfig.Application, DefaultSubnetSpec)
	}
}

// WithBitcoinSubnet adds an empty Bitcoin subnet.
func WithBitcoinSubnet() Option {
	return func(p *Config) {
		p.subnetConfig.Bitcoin = &DefaultSubnetSpec
	}
}

// WithDTSFlag sets the DTS flag for all subnets.
func WithDTSFlag() Option {
	return func(p *Config) {
		for _, subnet := range p.subnetConfig.Application {
			subnet.WithDTSFlag()
		}
		p.subnetConfig.Bitcoin.WithDTSFlag()
		p.subnetConfig.Fiduciary.WithDTSFlag()
		p.subnetConfig.II.WithDTSFlag()
		p.subnetConfig.NNS.WithDTSFlag()
		p.subnetConfig.SNS.WithDTSFlag()
		for _, subnet := range p.subnetConfig.System {
			subnet.WithDTSFlag()
		}
	}
}

// WithFiduciarySubnet adds an empty Fiduciary subnet.
func WithFiduciarySubnet() Option {
	return func(p *Config) {
		p.subnetConfig.Fiduciary = &DefaultSubnetSpec
	}
}

// WithHTTPClient sets the HTTP client for the PocketIC client.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Config) {
		p.client = client
	}
}

// WithIISubnet adds an empty Internet Identity subnet.
func WithIISubnet() Option {
	return func(p *Config) {
		p.subnetConfig.II = &DefaultSubnetSpec
	}
}

// WithLogger sets the logger for the PocketIC client.
func WithLogger(logger agent.Logger) Option {
	return func(p *Config) {
		p.logger = logger
	}
}

// WithNNSSubnet adds an empty NNS subnet.
func WithNNSSubnet() Option {
	return func(p *Config) {
		p.subnetConfig.NNS = &DefaultSubnetSpec
	}
}

func WithPollingDelay(delay, timeout time.Duration) Option {
	return func(p *Config) {
		p.delay = delay
		p.timeout = timeout
	}
}

// WithSNSSubnet adds an empty SNS subnet.
func WithSNSSubnet() Option {
	return func(p *Config) {
		p.subnetConfig.SNS = &DefaultSubnetSpec
	}
}

// WithSubnetConfigSet sets the subnet configuration.
func WithSubnetConfigSet(subnetConfig SubnetConfigSet) Option {
	return func(p *Config) {
		p.subnetConfig = subnetConfig
	}
}

// WithSystemSubnet adds an empty System subnet.
func WithSystemSubnet() Option {
	return func(p *Config) {
		p.subnetConfig.System = append(p.subnetConfig.System, DefaultSubnetSpec)
	}
}

// WithTTL sets the time-to-live for the PocketIC server, in seconds.
func WithTTL(ttl int) Option {
	return func(p *Config) {
		p.serverConfig = append(p.serverConfig, withTTL(ttl))
	}
}

// PocketIC is a client for the local PocketIC server.
type PocketIC struct {
	InstanceID  int
	httpGateway *HttpGatewayInfo
	topology    map[string]Topology

	logger         agent.Logger
	client         *http.Client
	delay, timeout time.Duration
	server         *server
}

// New creates a new PocketIC client.
// The order of the options is important, some options may override others.
func New(opts ...Option) (*PocketIC, error) {
	config := Config{
		subnetConfig: DefaultSubnetConfig,
		client:       http.DefaultClient,
		logger:       new(agent.NoopLogger),
		delay:        10 * time.Millisecond,
		timeout:      1 * time.Second,
	}
	for _, fn := range opts {
		fn(&config)
	}

	s, err := newServer(config.serverConfig...)
	if err != nil {
		return nil, err
	}

	// Create a new instance.
	req, err := newRequest(http.MethodPost, fmt.Sprintf("%s/instances", s.URL()), config.subnetConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}
	resp, err := config.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}
	var respBody createResponse[InstanceConfig]
	if respBody.Error != nil {
		return nil, respBody.Error
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create instance: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}

	return &PocketIC{
		InstanceID:  respBody.Created.InstanceID,
		httpGateway: nil,
		topology:    respBody.Created.Topology,
		logger:      config.logger,
		client:      config.client,
		delay:       config.delay,
		timeout:     config.timeout,
		server:      s,
	}, nil
}

// InstanceURL returns the URL of the PocketIC instance.
func (pic PocketIC) InstanceURL() string {
	return fmt.Sprintf("%s/instances/%d", pic.server.URL(), pic.InstanceID)
}

// Status pings the PocketIC instance.
func (pic PocketIC) Status() error {
	return pic.do(
		http.MethodGet,
		fmt.Sprintf("%s/status", pic.server.URL()),
		nil,
		nil,
	)
}

// Topology returns the topology of the PocketIC instance.
func (pic PocketIC) Topology() map[string]Topology {
	return pic.topology
}

// VerifySignature verifies a signature.
func (pic PocketIC) VerifySignature(sig VerifyCanisterSigArg) error {
	return pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/verify_signature", pic.server.URL()),
		sig,
		nil,
	)
}

type SubnetConfigSet struct {
	Application []SubnetSpec `json:"application"`
	Bitcoin     *SubnetSpec  `json:"bitcoin,omitempty"`
	Fiduciary   *SubnetSpec  `json:"fiduciary,omitempty"`
	II          *SubnetSpec  `json:"ii,omitempty"`
	NNS         *SubnetSpec  `json:"nns,omitempty"`
	SNS         *SubnetSpec  `json:"sns,omitempty"`
	System      []SubnetSpec `json:"system"`
}

type SubnetInstructionConfig interface {
	instructionConfig()
}

// SubnetInstructionConfigBenchmarking uses very high instruction limits useful for asymptotic canister benchmarking.
type SubnetInstructionConfigBenchmarking struct{}

func (c SubnetInstructionConfigBenchmarking) MarshalJSON() ([]byte, error) {
	return json.Marshal("Benchmarking")
}

func (c SubnetInstructionConfigBenchmarking) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	if s != "Benchmarking" {
		return fmt.Errorf("invalid instruction config: %s", s)
	}
	return nil
}

func (SubnetInstructionConfigBenchmarking) instructionConfig() {}

// SubnetInstructionConfigProduction uses default instruction limits as in production.
type SubnetInstructionConfigProduction struct{}

func (c SubnetInstructionConfigProduction) MarshalJSON() ([]byte, error) {
	return json.Marshal("Production")
}

func (c SubnetInstructionConfigProduction) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	if s != "Production" {
		return fmt.Errorf("invalid instruction config: %s", s)
	}
	return nil
}

func (SubnetInstructionConfigProduction) instructionConfig() {}

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

// SubnetSpec specifies various configurations for a subnet.
type SubnetSpec struct {
	StateConfig       SubnetStateConfig       `json:"state_config"`
	InstructionConfig SubnetInstructionConfig `json:"instruction_config"`
	DTSFlag           DTSFlag                 `json:"dts_flag"`
}

// WithDTSFlag sets the DTS flag, returns if the SubnetSpec is nil.
// Safe to call on a nil SubnetSpec.
func (s *SubnetSpec) WithDTSFlag() {
	if s == nil {
		return
	}
	s.DTSFlag = true
}

type SubnetStateConfig interface {
	stateConfig()
}

// SubnetStateConfigFromPath load existing subnet state from the given path. The path must be on a filesystem
// accessible to the server process.
type SubnetStateConfigFromPath struct {
	Path     string
	SubnetID SubnetID
}

func (c SubnetStateConfigFromPath) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{c.Path, c.SubnetID})
}

func (c SubnetStateConfigFromPath) UnmarshalJSON(bytes []byte) error {
	var v []json.RawMessage
	if err := json.Unmarshal(bytes, &v); err != nil {
		return err
	}
	if len(v) != 2 {
		return fmt.Errorf("invalid state config: %v", v)
	}
	if err := json.Unmarshal(v[0], &c.Path); err != nil {
		return err
	}
	return json.Unmarshal(v[1], &c.SubnetID)
}

func (SubnetStateConfigFromPath) stateConfig() {}

// SubnetStateConfigNew creates new subnet with empty state.
type SubnetStateConfigNew struct{}

func (c SubnetStateConfigNew) MarshalJSON() ([]byte, error) {
	return json.Marshal("New")
}

func (c SubnetStateConfigNew) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	if s != "New" {
		return fmt.Errorf("invalid state config: %s", s)
	}
	return nil
}

func (SubnetStateConfigNew) stateConfig() {}

type Topology struct {
	SubnetKind     SubnetKind        `json:"subnet_kind"`
	Size           int               `json:"size"`
	CanisterRanges []CanisterIDRange `json:"canister_ranges"`
}

type canisterIDRange struct {
	Start rawCanisterID `json:"start"`
	End   rawCanisterID `json:"end"`
}

type rawCanisterID struct {
	CanisterID string `json:"canister_id"`
}
