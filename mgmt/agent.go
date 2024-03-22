package mgmt

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

// Agent is a client for the "ic" canister.
type Agent struct {
	a          *agent.Agent
	canisterId principal.Principal
}

// NewAgent creates a new agent for the "ic" canister.
func NewAgent(config agent.Config) (*Agent, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}
	return &Agent{
		a:          a,
		canisterId: principal.Principal{}, // aaaaa-aa
	}, nil
}

// BitcoinGetBalance calls the "bitcoin_get_balance" method on the "ic" canister.
func (a Agent) BitcoinGetBalance(_ BitcoinGetBalanceArgs) (*BitcoinGetBalanceResult, error) {
	return nil, fmt.Errorf("bitcoin_get_balance is not accepted as an ingress message")
}

// BitcoinGetBalanceQuery calls the "bitcoin_get_balance_query" method on the "ic" canister.
func (a Agent) BitcoinGetBalanceQuery(arg0 BitcoinGetBalanceQueryArgs) (*BitcoinGetBalanceQueryResult, error) {
	var r0 BitcoinGetBalanceQueryResult
	if err := a.a.Query(
		a.canisterId,
		"bitcoin_get_balance_query",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// BitcoinGetCurrentFeePercentiles calls the "bitcoin_get_current_fee_percentiles" method on the "ic" canister.
func (a Agent) BitcoinGetCurrentFeePercentiles(_ BitcoinGetCurrentFeePercentilesArgs) (*BitcoinGetCurrentFeePercentilesResult, error) {
	return nil, fmt.Errorf("bitcoin_get_current_fee_percentiles is not accepted as an ingress message")
}

// BitcoinGetUtxos calls the "bitcoin_get_utxos" method on the "ic" canister.
func (a Agent) BitcoinGetUtxos(_ BitcoinGetUtxosArgs) (*BitcoinGetUtxosResult, error) {
	return nil, fmt.Errorf("bitcoin_get_utxos is not accepted as an ingress message")
}

// BitcoinGetUtxosQuery calls the "bitcoin_get_utxos_query" method on the "ic" canister.
func (a Agent) BitcoinGetUtxosQuery(arg0 BitcoinGetUtxosQueryArgs) (*BitcoinGetUtxosQueryResult, error) {
	var r0 BitcoinGetUtxosQueryResult
	if err := a.a.Query(
		a.canisterId,
		"bitcoin_get_utxos_query",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// BitcoinSendTransaction calls the "bitcoin_send_transaction" method on the "ic" canister.
func (a Agent) BitcoinSendTransaction(_ BitcoinSendTransactionArgs) error {
	return fmt.Errorf("bitcoin_send_transaction is not accepted as an ingress message")
}

// CanisterInfo calls the "canister_info" method on the "ic" canister.
func (a Agent) CanisterInfo(arg0 CanisterInfoArgs) (*CanisterInfoResult, error) {
	var r0 CanisterInfoResult
	if err := a.a.Call(
		a.canisterId,
		"canister_info",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// CanisterStatus calls the "canister_status" method on the "ic" canister.
func (a Agent) CanisterStatus(arg0 CanisterStatusArgs) (*CanisterStatusResult, error) {
	var r0 CanisterStatusResult
	if err := a.a.Call(
		arg0.CanisterId,
		"canister_status",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// ClearChunkStore calls the "clear_chunk_store" method on the "ic" canister.
func (a Agent) ClearChunkStore(arg0 ClearChunkStoreArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"clear_chunk_store",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// CreateCanister calls the "create_canister" method on the "ic" canister.
func (a Agent) CreateCanister(_ CreateCanisterArgs) (*CreateCanisterResult, error) {
	return nil, fmt.Errorf("create_canister is not accepted as an ingress message")
}

// DeleteCanister calls the "delete_canister" method on the "ic" canister.
func (a Agent) DeleteCanister(arg0 DeleteCanisterArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"delete_canister",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// DepositCycles calls the "deposit_cycles" method on the "ic" canister.
func (a Agent) DepositCycles(arg0 DepositCyclesArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"deposit_cycles",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// EcdsaPublicKey calls the "ecdsa_public_key" method on the "ic" canister.
func (a Agent) EcdsaPublicKey(_ EcdsaPublicKeyArgs) (*EcdsaPublicKeyResult, error) {
	return nil, fmt.Errorf("ecdsa_public_key is not accepted as an ingress message")
}

// HttpRequest calls the "http_request" method on the "ic" canister.
func (a Agent) HttpRequest(arg0 HttpRequestArgs) (*HttpRequestResult, error) {
	var r0 HttpRequestResult
	if err := a.a.Call(
		a.canisterId,
		"http_request",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// InstallChunkedCode calls the "install_chunked_code" method on the "ic" canister.
func (a Agent) InstallChunkedCode(arg0 InstallChunkedCodeArgs) error {
	if err := a.a.Call(
		arg0.TargetCanister,
		"install_chunked_code",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// InstallCode calls the "install_code" method on the "ic" canister.
func (a Agent) InstallCode(arg0 InstallCodeArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"install_code",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// NodeMetricsHistory calls the "node_metrics_history" method on the "ic" canister.
func (a Agent) NodeMetricsHistory(_ NodeMetricsHistoryArgs) (*NodeMetricsHistoryResult, error) {
	return nil, fmt.Errorf("node_metrics_history is not accepted as an ingress message")
}

// ProvisionalCreateCanisterWithCycles calls the "provisional_create_canister_with_cycles" method on the "ic" canister.
func (a Agent) ProvisionalCreateCanisterWithCycles(arg0 ProvisionalCreateCanisterWithCyclesArgs) (*ProvisionalCreateCanisterWithCyclesResult, error) {
	var r0 ProvisionalCreateCanisterWithCyclesResult
	if err := a.a.Call(
		a.canisterId,
		"provisional_create_canister_with_cycles",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// ProvisionalTopUpCanister calls the "provisional_top_up_canister" method on the "ic" canister.
func (a Agent) ProvisionalTopUpCanister(arg0 ProvisionalTopUpCanisterArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"provisional_top_up_canister",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// RawRand calls the "raw_rand" method on the "ic" canister.
func (a Agent) RawRand() (*RawRandResult, error) {
	return nil, fmt.Errorf("raw_rand is not accepted as an ingress message")
}

// SignWithEcdsa calls the "sign_with_ecdsa" method on the "ic" canister.
func (a Agent) SignWithEcdsa(_ SignWithEcdsaArgs) (*SignWithEcdsaResult, error) {
	return nil, fmt.Errorf("sign_with_ecdsa is not accepted as an ingress message")
}

// StartCanister calls the "start_canister" method on the "ic" canister.
func (a Agent) StartCanister(arg0 StartCanisterArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"start_canister",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// StopCanister calls the "stop_canister" method on the "ic" canister.
func (a Agent) StopCanister(arg0 StopCanisterArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"stop_canister",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// StoredChunks calls the "stored_chunks" method on the "ic" canister.
func (a Agent) StoredChunks(arg0 StoredChunksArgs) (*StoredChunksResult, error) {
	var r0 StoredChunksResult
	if err := a.a.Call(
		arg0.CanisterId,
		"stored_chunks",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// UninstallCode calls the "uninstall_code" method on the "ic" canister.
func (a Agent) UninstallCode(arg0 UninstallCodeArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"uninstall_code",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// UpdateSettings calls the "update_settings" method on the "ic" canister.
func (a Agent) UpdateSettings(arg0 UpdateSettingsArgs) error {
	if err := a.a.Call(
		arg0.CanisterId,
		"update_settings",
		[]any{arg0},
		[]any{},
	); err != nil {
		return err
	}
	return nil
}

// UploadChunk calls the "upload_chunk" method on the "ic" canister.
func (a Agent) UploadChunk(arg0 UploadChunkArgs) (*UploadChunkResult, error) {
	var r0 UploadChunkResult
	if err := a.a.Call(
		arg0.CanisterId,
		"upload_chunk",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

type BitcoinAddress = string

type BitcoinGetBalanceArgs struct {
	Address          BitcoinAddress `ic:"address" json:"address"`
	Network          BitcoinNetwork `ic:"network" json:"network"`
	MinConfirmations *uint32        `ic:"min_confirmations,omitempty" json:"min_confirmations,omitempty"`
}

type BitcoinGetBalanceQueryArgs struct {
	Address          BitcoinAddress `ic:"address" json:"address"`
	Network          BitcoinNetwork `ic:"network" json:"network"`
	MinConfirmations *uint32        `ic:"min_confirmations,omitempty" json:"min_confirmations,omitempty"`
}

type BitcoinGetBalanceQueryResult = Satoshi

type BitcoinGetBalanceResult = Satoshi

type BitcoinGetCurrentFeePercentilesArgs struct {
	Network BitcoinNetwork `ic:"network" json:"network"`
}

type BitcoinGetCurrentFeePercentilesResult = []MillisatoshiPerByte

type BitcoinGetUtxosArgs struct {
	Address BitcoinAddress `ic:"address" json:"address"`
	Network BitcoinNetwork `ic:"network" json:"network"`
	Filter  *struct {
		MinConfirmations *uint32 `ic:"min_confirmations,variant"`
		Page             *[]byte `ic:"page,variant"`
	} `ic:"filter,omitempty" json:"filter,omitempty"`
}

type BitcoinGetUtxosQueryArgs struct {
	Address BitcoinAddress `ic:"address" json:"address"`
	Network BitcoinNetwork `ic:"network" json:"network"`
	Filter  *struct {
		MinConfirmations *uint32 `ic:"min_confirmations,variant"`
		Page             *[]byte `ic:"page,variant"`
	} `ic:"filter,omitempty" json:"filter,omitempty"`
}

type BitcoinGetUtxosQueryResult struct {
	Utxos        []Utxo    `ic:"utxos" json:"utxos"`
	TipBlockHash BlockHash `ic:"tip_block_hash" json:"tip_block_hash"`
	TipHeight    uint32    `ic:"tip_height" json:"tip_height"`
	NextPage     *[]byte   `ic:"next_page,omitempty" json:"next_page,omitempty"`
}

type BitcoinGetUtxosResult struct {
	Utxos        []Utxo    `ic:"utxos" json:"utxos"`
	TipBlockHash BlockHash `ic:"tip_block_hash" json:"tip_block_hash"`
	TipHeight    uint32    `ic:"tip_height" json:"tip_height"`
	NextPage     *[]byte   `ic:"next_page,omitempty" json:"next_page,omitempty"`
}

type BitcoinNetwork struct {
	Mainnet *idl.Null `ic:"mainnet,variant"`
	Testnet *idl.Null `ic:"testnet,variant"`
}

type BitcoinSendTransactionArgs struct {
	Transaction []byte         `ic:"transaction" json:"transaction"`
	Network     BitcoinNetwork `ic:"network" json:"network"`
}

type BlockHash = []byte

type CanisterId = principal.Principal

type CanisterInfoArgs struct {
	CanisterId          CanisterId `ic:"canister_id" json:"canister_id"`
	NumRequestedChanges *uint64    `ic:"num_requested_changes,omitempty" json:"num_requested_changes,omitempty"`
}

type CanisterInfoResult struct {
	TotalNumChanges uint64                `ic:"total_num_changes" json:"total_num_changes"`
	RecentChanges   []Change              `ic:"recent_changes" json:"recent_changes"`
	ModuleHash      *[]byte               `ic:"module_hash,omitempty" json:"module_hash,omitempty"`
	Controllers     []principal.Principal `ic:"controllers" json:"controllers"`
}

type CanisterInstallMode struct {
	Install   *idl.Null `ic:"install,variant"`
	Reinstall *idl.Null `ic:"reinstall,variant"`
	Upgrade   **struct {
		SkipPreUpgrade *bool `ic:"skip_pre_upgrade,omitempty" json:"skip_pre_upgrade,omitempty"`
	} `ic:"upgrade,variant"`
}

type CanisterSettings struct {
	Controllers         *[]principal.Principal `ic:"controllers,omitempty" json:"controllers,omitempty"`
	ComputeAllocation   *idl.Nat               `ic:"compute_allocation,omitempty" json:"compute_allocation,omitempty"`
	MemoryAllocation    *idl.Nat               `ic:"memory_allocation,omitempty" json:"memory_allocation,omitempty"`
	FreezingThreshold   *idl.Nat               `ic:"freezing_threshold,omitempty" json:"freezing_threshold,omitempty"`
	ReservedCyclesLimit *idl.Nat               `ic:"reserved_cycles_limit,omitempty" json:"reserved_cycles_limit,omitempty"`
}

type CanisterStatusArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type CanisterStatusResult struct {
	Status struct {
		Running  *idl.Null `ic:"running,variant"`
		Stopping *idl.Null `ic:"stopping,variant"`
		Stopped  *idl.Null `ic:"stopped,variant"`
	} `ic:"status" json:"status"`
	Settings               DefiniteCanisterSettings `ic:"settings" json:"settings"`
	ModuleHash             *[]byte                  `ic:"module_hash,omitempty" json:"module_hash,omitempty"`
	MemorySize             idl.Nat                  `ic:"memory_size" json:"memory_size"`
	Cycles                 idl.Nat                  `ic:"cycles" json:"cycles"`
	ReservedCycles         idl.Nat                  `ic:"reserved_cycles" json:"reserved_cycles"`
	IdleCyclesBurnedPerDay idl.Nat                  `ic:"idle_cycles_burned_per_day" json:"idle_cycles_burned_per_day"`
}

type Change struct {
	TimestampNanos  uint64        `ic:"timestamp_nanos" json:"timestamp_nanos"`
	CanisterVersion uint64        `ic:"canister_version" json:"canister_version"`
	Origin          ChangeOrigin  `ic:"origin" json:"origin"`
	Details         ChangeDetails `ic:"details" json:"details"`
}

type ChangeDetails struct {
	Creation *struct {
		Controllers []principal.Principal `ic:"controllers" json:"controllers"`
	} `ic:"creation,variant"`
	CodeUninstall  *idl.Null `ic:"code_uninstall,variant"`
	CodeDeployment *struct {
		Mode struct {
			Install   *idl.Null `ic:"install,variant"`
			Reinstall *idl.Null `ic:"reinstall,variant"`
			Upgrade   *idl.Null `ic:"upgrade,variant"`
		} `ic:"mode" json:"mode"`
		ModuleHash []byte `ic:"module_hash" json:"module_hash"`
	} `ic:"code_deployment,variant"`
	ControllersChange *struct {
		Controllers []principal.Principal `ic:"controllers" json:"controllers"`
	} `ic:"controllers_change,variant"`
}

type ChangeOrigin struct {
	FromUser *struct {
		UserId principal.Principal `ic:"user_id" json:"user_id"`
	} `ic:"from_user,variant"`
	FromCanister *struct {
		CanisterId      principal.Principal `ic:"canister_id" json:"canister_id"`
		CanisterVersion *uint64             `ic:"canister_version,omitempty" json:"canister_version,omitempty"`
	} `ic:"from_canister,variant"`
}

type ChunkHash struct {
	Hash []byte `ic:"hash" json:"hash"`
}

type ClearChunkStoreArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type CreateCanisterArgs struct {
	Settings              *CanisterSettings `ic:"settings,omitempty" json:"settings,omitempty"`
	SenderCanisterVersion *uint64           `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type CreateCanisterResult struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type DefiniteCanisterSettings struct {
	Controllers         []principal.Principal `ic:"controllers" json:"controllers"`
	ComputeAllocation   idl.Nat               `ic:"compute_allocation" json:"compute_allocation"`
	MemoryAllocation    idl.Nat               `ic:"memory_allocation" json:"memory_allocation"`
	FreezingThreshold   idl.Nat               `ic:"freezing_threshold" json:"freezing_threshold"`
	ReservedCyclesLimit idl.Nat               `ic:"reserved_cycles_limit" json:"reserved_cycles_limit"`
}

type DeleteCanisterArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type DepositCyclesArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type EcdsaCurve struct {
	Secp256k1 *idl.Null `ic:"secp256k1,variant"`
}

type EcdsaPublicKeyArgs struct {
	CanisterId     *CanisterId `ic:"canister_id,omitempty" json:"canister_id,omitempty"`
	DerivationPath [][]byte    `ic:"derivation_path" json:"derivation_path"`
	KeyId          struct {
		Curve EcdsaCurve `ic:"curve" json:"curve"`
		Name  string     `ic:"name" json:"name"`
	} `ic:"key_id" json:"key_id"`
}

type EcdsaPublicKeyResult struct {
	PublicKey []byte `ic:"public_key" json:"public_key"`
	ChainCode []byte `ic:"chain_code" json:"chain_code"`
}

type HttpHeader struct {
	Name  string `ic:"name" json:"name"`
	Value string `ic:"value" json:"value"`
}

type HttpRequestArgs struct {
	Url              string  `ic:"url" json:"url"`
	MaxResponseBytes *uint64 `ic:"max_response_bytes,omitempty" json:"max_response_bytes,omitempty"`
	Method           struct {
		Get  *idl.Null `ic:"get,variant"`
		Head *idl.Null `ic:"head,variant"`
		Post *idl.Null `ic:"post,variant"`
	} `ic:"method" json:"method"`
	Headers   []HttpHeader `ic:"headers" json:"headers"`
	Body      *[]byte      `ic:"body,omitempty" json:"body,omitempty"`
	Transform *struct {
		Function struct { /* NOT SUPPORTED */
		} `ic:"function" json:"function"`
		Context []byte `ic:"context" json:"context"`
	} `ic:"transform,omitempty" json:"transform,omitempty"`
}

type HttpRequestResult struct {
	Status  idl.Nat      `ic:"status" json:"status"`
	Headers []HttpHeader `ic:"headers" json:"headers"`
	Body    []byte       `ic:"body" json:"body"`
}

type InstallChunkedCodeArgs struct {
	Mode                  CanisterInstallMode `ic:"mode" json:"mode"`
	TargetCanister        CanisterId          `ic:"target_canister" json:"target_canister"`
	StoreCanister         *CanisterId         `ic:"store_canister,omitempty" json:"store_canister,omitempty"`
	ChunkHashesList       []ChunkHash         `ic:"chunk_hashes_list" json:"chunk_hashes_list"`
	WasmModuleHash        []byte              `ic:"wasm_module_hash" json:"wasm_module_hash"`
	Arg                   []byte              `ic:"arg" json:"arg"`
	SenderCanisterVersion *uint64             `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type InstallCodeArgs struct {
	Mode                  CanisterInstallMode `ic:"mode" json:"mode"`
	CanisterId            CanisterId          `ic:"canister_id" json:"canister_id"`
	WasmModule            WasmModule          `ic:"wasm_module" json:"wasm_module"`
	Arg                   []byte              `ic:"arg" json:"arg"`
	SenderCanisterVersion *uint64             `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type MillisatoshiPerByte = uint64

type NodeMetrics struct {
	NodeId                principal.Principal `ic:"node_id" json:"node_id"`
	NumBlocksTotal        uint64              `ic:"num_blocks_total" json:"num_blocks_total"`
	NumBlockFailuresTotal uint64              `ic:"num_block_failures_total" json:"num_block_failures_total"`
}

type NodeMetricsHistoryArgs struct {
	SubnetId              principal.Principal `ic:"subnet_id" json:"subnet_id"`
	StartAtTimestampNanos uint64              `ic:"start_at_timestamp_nanos" json:"start_at_timestamp_nanos"`
}

type NodeMetricsHistoryResult = []struct {
	TimestampNanos uint64        `ic:"timestamp_nanos" json:"timestamp_nanos"`
	NodeMetrics    []NodeMetrics `ic:"node_metrics" json:"node_metrics"`
}

type Outpoint struct {
	Txid []byte `ic:"txid" json:"txid"`
	Vout uint32 `ic:"vout" json:"vout"`
}

type ProvisionalCreateCanisterWithCyclesArgs struct {
	Amount                *idl.Nat          `ic:"amount,omitempty" json:"amount,omitempty"`
	Settings              *CanisterSettings `ic:"settings,omitempty" json:"settings,omitempty"`
	SpecifiedId           *CanisterId       `ic:"specified_id,omitempty" json:"specified_id,omitempty"`
	SenderCanisterVersion *uint64           `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type ProvisionalCreateCanisterWithCyclesResult struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type ProvisionalTopUpCanisterArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
	Amount     idl.Nat    `ic:"amount" json:"amount"`
}

type RawRandResult = []byte

type Satoshi = uint64

type SignWithEcdsaArgs struct {
	MessageHash    []byte   `ic:"message_hash" json:"message_hash"`
	DerivationPath [][]byte `ic:"derivation_path" json:"derivation_path"`
	KeyId          struct {
		Curve EcdsaCurve `ic:"curve" json:"curve"`
		Name  string     `ic:"name" json:"name"`
	} `ic:"key_id" json:"key_id"`
}

type SignWithEcdsaResult struct {
	Signature []byte `ic:"signature" json:"signature"`
}

type StartCanisterArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type StopCanisterArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type StoredChunksArgs struct {
	CanisterId CanisterId `ic:"canister_id" json:"canister_id"`
}

type StoredChunksResult = []ChunkHash

type UninstallCodeArgs struct {
	CanisterId            CanisterId `ic:"canister_id" json:"canister_id"`
	SenderCanisterVersion *uint64    `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type UpdateSettingsArgs struct {
	CanisterId            principal.Principal `ic:"canister_id" json:"canister_id"`
	Settings              CanisterSettings    `ic:"settings" json:"settings"`
	SenderCanisterVersion *uint64             `ic:"sender_canister_version,omitempty" json:"sender_canister_version,omitempty"`
}

type UploadChunkArgs struct {
	CanisterId principal.Principal `ic:"canister_id" json:"canister_id"`
	Chunk      []byte              `ic:"chunk" json:"chunk"`
}

type UploadChunkResult = ChunkHash

type Utxo struct {
	Outpoint Outpoint `ic:"outpoint" json:"outpoint"`
	Value    Satoshi  `ic:"value" json:"value"`
	Height   uint32   `ic:"height" json:"height"`
}

type WasmModule = []byte
