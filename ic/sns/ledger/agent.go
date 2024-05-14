// Package ledger provides a client for the "ledger" canister.
// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
package ledger

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

type Account struct {
	Owner      principal.Principal `ic:"owner" json:"owner"`
	Subaccount *Subaccount         `ic:"subaccount,omitempty" json:"subaccount,omitempty"`
}

// Agent is a client for the "ledger" canister.
type Agent struct {
	a          *agent.Agent
	canisterId principal.Principal
}

// NewAgent creates a new agent for the "ledger" canister.
func NewAgent(canisterId principal.Principal, config agent.Config) (*Agent, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}
	return &Agent{
		a:          a,
		canisterId: canisterId,
	}, nil
}

// Archives calls the "archives" method on the "ledger" canister.
func (a Agent) Archives() (*[]ArchiveInfo, error) {
	var r0 []ArchiveInfo
	if err := a.a.Query(
		a.canisterId,
		"archives",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// GetBlocks calls the "get_blocks" method on the "ledger" canister.
func (a Agent) GetBlocks(arg0 GetBlocksArgs) (*GetBlocksResponse, error) {
	var r0 GetBlocksResponse
	if err := a.a.Query(
		a.canisterId,
		"get_blocks",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// GetDataCertificate calls the "get_data_certificate" method on the "ledger" canister.
func (a Agent) GetDataCertificate() (*DataCertificate, error) {
	var r0 DataCertificate
	if err := a.a.Query(
		a.canisterId,
		"get_data_certificate",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// GetTransactions calls the "get_transactions" method on the "ledger" canister.
func (a Agent) GetTransactions(arg0 GetTransactionsRequest) (*GetTransactionsResponse, error) {
	var r0 GetTransactionsResponse
	if err := a.a.Query(
		a.canisterId,
		"get_transactions",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1BalanceOf calls the "icrc1_balance_of" method on the "ledger" canister.
func (a Agent) Icrc1BalanceOf(arg0 Account) (*Tokens, error) {
	var r0 Tokens
	if err := a.a.Query(
		a.canisterId,
		"icrc1_balance_of",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Decimals calls the "icrc1_decimals" method on the "ledger" canister.
func (a Agent) Icrc1Decimals() (*uint8, error) {
	var r0 uint8
	if err := a.a.Query(
		a.canisterId,
		"icrc1_decimals",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Fee calls the "icrc1_fee" method on the "ledger" canister.
func (a Agent) Icrc1Fee() (*Tokens, error) {
	var r0 Tokens
	if err := a.a.Query(
		a.canisterId,
		"icrc1_fee",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Metadata calls the "icrc1_metadata" method on the "ledger" canister.
func (a Agent) Icrc1Metadata() (*[]struct {
	Field0 string        `ic:"0" json:"0"`
	Field1 MetadataValue `ic:"1" json:"1"`
}, error) {
	var r0 []struct {
		Field0 string        `ic:"0" json:"0"`
		Field1 MetadataValue `ic:"1" json:"1"`
	}
	if err := a.a.Query(
		a.canisterId,
		"icrc1_metadata",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1MintingAccount calls the "icrc1_minting_account" method on the "ledger" canister.
func (a Agent) Icrc1MintingAccount() (**Account, error) {
	var r0 *Account
	if err := a.a.Query(
		a.canisterId,
		"icrc1_minting_account",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Name calls the "icrc1_name" method on the "ledger" canister.
func (a Agent) Icrc1Name() (*string, error) {
	var r0 string
	if err := a.a.Query(
		a.canisterId,
		"icrc1_name",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1SupportedStandards calls the "icrc1_supported_standards" method on the "ledger" canister.
func (a Agent) Icrc1SupportedStandards() (*[]StandardRecord, error) {
	var r0 []StandardRecord
	if err := a.a.Query(
		a.canisterId,
		"icrc1_supported_standards",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Symbol calls the "icrc1_symbol" method on the "ledger" canister.
func (a Agent) Icrc1Symbol() (*string, error) {
	var r0 string
	if err := a.a.Query(
		a.canisterId,
		"icrc1_symbol",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1TotalSupply calls the "icrc1_total_supply" method on the "ledger" canister.
func (a Agent) Icrc1TotalSupply() (*Tokens, error) {
	var r0 Tokens
	if err := a.a.Query(
		a.canisterId,
		"icrc1_total_supply",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc1Transfer calls the "icrc1_transfer" method on the "ledger" canister.
func (a Agent) Icrc1Transfer(arg0 TransferArg) (*TransferResult, error) {
	var r0 TransferResult
	if err := a.a.Call(
		a.canisterId,
		"icrc1_transfer",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc2Allowance calls the "icrc2_allowance" method on the "ledger" canister.
func (a Agent) Icrc2Allowance(arg0 AllowanceArgs) (*Allowance, error) {
	var r0 Allowance
	if err := a.a.Query(
		a.canisterId,
		"icrc2_allowance",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc2Approve calls the "icrc2_approve" method on the "ledger" canister.
func (a Agent) Icrc2Approve(arg0 ApproveArgs) (*ApproveResult, error) {
	var r0 ApproveResult
	if err := a.a.Call(
		a.canisterId,
		"icrc2_approve",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc2TransferFrom calls the "icrc2_transfer_from" method on the "ledger" canister.
func (a Agent) Icrc2TransferFrom(arg0 TransferFromArgs) (*TransferFromResult, error) {
	var r0 TransferFromResult
	if err := a.a.Call(
		a.canisterId,
		"icrc2_transfer_from",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc3GetArchives calls the "icrc3_get_archives" method on the "ledger" canister.
func (a Agent) Icrc3GetArchives(arg0 GetArchivesArgs) (*GetArchivesResult, error) {
	var r0 GetArchivesResult
	if err := a.a.Query(
		a.canisterId,
		"icrc3_get_archives",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc3GetBlocks calls the "icrc3_get_blocks" method on the "ledger" canister.
func (a Agent) Icrc3GetBlocks(arg0 []GetBlocksArgs) (*GetBlocksResult, error) {
	var r0 GetBlocksResult
	if err := a.a.Query(
		a.canisterId,
		"icrc3_get_blocks",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc3GetTipCertificate calls the "icrc3_get_tip_certificate" method on the "ledger" canister.
func (a Agent) Icrc3GetTipCertificate() (**ICRC3DataCertificate, error) {
	var r0 *ICRC3DataCertificate
	if err := a.a.Query(
		a.canisterId,
		"icrc3_get_tip_certificate",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// Icrc3SupportedBlockTypes calls the "icrc3_supported_block_types" method on the "ledger" canister.
func (a Agent) Icrc3SupportedBlockTypes() (*[]struct {
	BlockType string `ic:"block_type" json:"block_type"`
	Url       string `ic:"url" json:"url"`
}, error) {
	var r0 []struct {
		BlockType string `ic:"block_type" json:"block_type"`
		Url       string `ic:"url" json:"url"`
	}
	if err := a.a.Query(
		a.canisterId,
		"icrc3_supported_block_types",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

type Allowance struct {
	Allowance idl.Nat    `ic:"allowance" json:"allowance"`
	ExpiresAt *Timestamp `ic:"expires_at,omitempty" json:"expires_at,omitempty"`
}

type AllowanceArgs struct {
	Account Account `ic:"account" json:"account"`
	Spender Account `ic:"spender" json:"spender"`
}

type Approve struct {
	Fee               *idl.Nat   `ic:"fee,omitempty" json:"fee,omitempty"`
	From              Account    `ic:"from" json:"from"`
	Memo              *[]byte    `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime     *Timestamp `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount            idl.Nat    `ic:"amount" json:"amount"`
	ExpectedAllowance *idl.Nat   `ic:"expected_allowance,omitempty" json:"expected_allowance,omitempty"`
	ExpiresAt         *Timestamp `ic:"expires_at,omitempty" json:"expires_at,omitempty"`
	Spender           Account    `ic:"spender" json:"spender"`
}

type ApproveArgs struct {
	Fee               *idl.Nat   `ic:"fee,omitempty" json:"fee,omitempty"`
	Memo              *[]byte    `ic:"memo,omitempty" json:"memo,omitempty"`
	FromSubaccount    *[]byte    `ic:"from_subaccount,omitempty" json:"from_subaccount,omitempty"`
	CreatedAtTime     *Timestamp `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount            idl.Nat    `ic:"amount" json:"amount"`
	ExpectedAllowance *idl.Nat   `ic:"expected_allowance,omitempty" json:"expected_allowance,omitempty"`
	ExpiresAt         *Timestamp `ic:"expires_at,omitempty" json:"expires_at,omitempty"`
	Spender           Account    `ic:"spender" json:"spender"`
}

type ApproveError struct {
	GenericError *struct {
		Message   string  `ic:"message" json:"message"`
		ErrorCode idl.Nat `ic:"error_code" json:"error_code"`
	} `ic:"GenericError,variant"`
	TemporarilyUnavailable *idl.Null `ic:"TemporarilyUnavailable,variant"`
	Duplicate              *struct {
		DuplicateOf BlockIndex `ic:"duplicate_of" json:"duplicate_of"`
	} `ic:"Duplicate,variant"`
	BadFee *struct {
		ExpectedFee idl.Nat `ic:"expected_fee" json:"expected_fee"`
	} `ic:"BadFee,variant"`
	AllowanceChanged *struct {
		CurrentAllowance idl.Nat `ic:"current_allowance" json:"current_allowance"`
	} `ic:"AllowanceChanged,variant"`
	CreatedInFuture *struct {
		LedgerTime Timestamp `ic:"ledger_time" json:"ledger_time"`
	} `ic:"CreatedInFuture,variant"`
	TooOld  *idl.Null `ic:"TooOld,variant"`
	Expired *struct {
		LedgerTime Timestamp `ic:"ledger_time" json:"ledger_time"`
	} `ic:"Expired,variant"`
	InsufficientFunds *struct {
		Balance idl.Nat `ic:"balance" json:"balance"`
	} `ic:"InsufficientFunds,variant"`
}

type ApproveResult struct {
	Ok  *BlockIndex   `ic:"Ok,variant"`
	Err *ApproveError `ic:"Err,variant"`
}

type ArchiveInfo struct {
	CanisterId      principal.Principal `ic:"canister_id" json:"canister_id"`
	BlockRangeStart BlockIndex          `ic:"block_range_start" json:"block_range_start"`
	BlockRangeEnd   BlockIndex          `ic:"block_range_end" json:"block_range_end"`
}

type Block = Value

type BlockIndex = idl.Nat

type BlockRange struct {
	Blocks []Block `ic:"blocks" json:"blocks"`
}

type Burn struct {
	From          Account    `ic:"from" json:"from"`
	Memo          *[]byte    `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *Timestamp `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat    `ic:"amount" json:"amount"`
	Spender       *Account   `ic:"spender,omitempty" json:"spender,omitempty"`
}

type ChangeArchiveOptions struct {
	NumBlocksToArchive         *uint64                `ic:"num_blocks_to_archive,omitempty" json:"num_blocks_to_archive,omitempty"`
	MaxTransactionsPerResponse *uint64                `ic:"max_transactions_per_response,omitempty" json:"max_transactions_per_response,omitempty"`
	TriggerThreshold           *uint64                `ic:"trigger_threshold,omitempty" json:"trigger_threshold,omitempty"`
	MaxMessageSizeBytes        *uint64                `ic:"max_message_size_bytes,omitempty" json:"max_message_size_bytes,omitempty"`
	CyclesForArchiveCreation   *uint64                `ic:"cycles_for_archive_creation,omitempty" json:"cycles_for_archive_creation,omitempty"`
	NodeMaxMemorySizeBytes     *uint64                `ic:"node_max_memory_size_bytes,omitempty" json:"node_max_memory_size_bytes,omitempty"`
	ControllerId               *principal.Principal   `ic:"controller_id,omitempty" json:"controller_id,omitempty"`
	MoreControllerIds          *[]principal.Principal `ic:"more_controller_ids,omitempty" json:"more_controller_ids,omitempty"`
}

type ChangeFeeCollector struct {
	Unset *idl.Null `ic:"Unset,variant"`
	SetTo *Account  `ic:"SetTo,variant"`
}

type DataCertificate struct {
	Certificate *[]byte `ic:"certificate,omitempty" json:"certificate,omitempty"`
	HashTree    []byte  `ic:"hash_tree" json:"hash_tree"`
}

type Duration = uint64

type FeatureFlags struct {
	Icrc2 bool `ic:"icrc2" json:"icrc2"`
}

type GetArchivesArgs struct {
	From *principal.Principal `ic:"from,omitempty" json:"from,omitempty"`
}

type GetArchivesResult = []struct {
	CanisterId principal.Principal `ic:"canister_id" json:"canister_id"`
	Start      idl.Nat             `ic:"start" json:"start"`
	End        idl.Nat             `ic:"end" json:"end"`
}

type GetBlocksArgs struct {
	Start  BlockIndex `ic:"start" json:"start"`
	Length idl.Nat    `ic:"length" json:"length"`
}

type GetBlocksResponse struct {
	FirstIndex     BlockIndex `ic:"first_index" json:"first_index"`
	ChainLength    uint64     `ic:"chain_length" json:"chain_length"`
	Certificate    *[]byte    `ic:"certificate,omitempty" json:"certificate,omitempty"`
	Blocks         []Block    `ic:"blocks" json:"blocks"`
	ArchivedBlocks []struct {
		Start    BlockIndex          `ic:"start" json:"start"`
		Length   idl.Nat             `ic:"length" json:"length"`
		Callback QueryBlockArchiveFn `ic:"callback" json:"callback"`
	} `ic:"archived_blocks" json:"archived_blocks"`
}

type GetBlocksResult struct {
	LogLength idl.Nat `ic:"log_length" json:"log_length"`
	Blocks    []struct {
		Id    idl.Nat    `ic:"id" json:"id"`
		Block ICRC3Value `ic:"block" json:"block"`
	} `ic:"blocks" json:"blocks"`
	ArchivedBlocks []struct {
		Args     []GetBlocksArgs `ic:"args" json:"args"`
		Callback struct { /* NOT SUPPORTED */
		} `ic:"callback" json:"callback"`
	} `ic:"archived_blocks" json:"archived_blocks"`
}

type GetTransactionsRequest struct {
	Start  TxIndex `ic:"start" json:"start"`
	Length idl.Nat `ic:"length" json:"length"`
}

type GetTransactionsResponse struct {
	LogLength            idl.Nat       `ic:"log_length" json:"log_length"`
	Transactions         []Transaction `ic:"transactions" json:"transactions"`
	FirstIndex           TxIndex       `ic:"first_index" json:"first_index"`
	ArchivedTransactions []struct {
		Start    TxIndex        `ic:"start" json:"start"`
		Length   idl.Nat        `ic:"length" json:"length"`
		Callback QueryArchiveFn `ic:"callback" json:"callback"`
	} `ic:"archived_transactions" json:"archived_transactions"`
}

type HttpRequest struct {
	Url     string `ic:"url" json:"url"`
	Method  string `ic:"method" json:"method"`
	Body    []byte `ic:"body" json:"body"`
	Headers []struct {
		Field0 string `ic:"0" json:"0"`
		Field1 string `ic:"1" json:"1"`
	} `ic:"headers" json:"headers"`
}

type HttpResponse struct {
	Body    []byte `ic:"body" json:"body"`
	Headers []struct {
		Field0 string `ic:"0" json:"0"`
		Field1 string `ic:"1" json:"1"`
	} `ic:"headers" json:"headers"`
	StatusCode uint16 `ic:"status_code" json:"status_code"`
}

type ICRC3DataCertificate struct {
	Certificate []byte `ic:"certificate" json:"certificate"`
	HashTree    []byte `ic:"hash_tree" json:"hash_tree"`
}

type ICRC3Value struct {
	Blob  *[]byte       `ic:"Blob,variant"`
	Text  *string       `ic:"Text,variant"`
	Nat   *idl.Nat      `ic:"Nat,variant"`
	Int   *idl.Int      `ic:"Int,variant"`
	Array *[]ICRC3Value `ic:"Array,variant"`
	Map   *[]struct {
		Field0 string     `ic:"0" json:"0"`
		Field1 ICRC3Value `ic:"1" json:"1"`
	} `ic:"Map,variant"`
}

type InitArgs struct {
	MintingAccount      Account  `ic:"minting_account" json:"minting_account"`
	FeeCollectorAccount *Account `ic:"fee_collector_account,omitempty" json:"fee_collector_account,omitempty"`
	TransferFee         idl.Nat  `ic:"transfer_fee" json:"transfer_fee"`
	Decimals            *uint8   `ic:"decimals,omitempty" json:"decimals,omitempty"`
	MaxMemoLength       *uint16  `ic:"max_memo_length,omitempty" json:"max_memo_length,omitempty"`
	TokenSymbol         string   `ic:"token_symbol" json:"token_symbol"`
	TokenName           string   `ic:"token_name" json:"token_name"`
	Metadata            []struct {
		Field0 string        `ic:"0" json:"0"`
		Field1 MetadataValue `ic:"1" json:"1"`
	} `ic:"metadata" json:"metadata"`
	InitialBalances []struct {
		Field0 Account `ic:"0" json:"0"`
		Field1 idl.Nat `ic:"1" json:"1"`
	} `ic:"initial_balances" json:"initial_balances"`
	FeatureFlags                 *FeatureFlags `ic:"feature_flags,omitempty" json:"feature_flags,omitempty"`
	MaximumNumberOfAccounts      *uint64       `ic:"maximum_number_of_accounts,omitempty" json:"maximum_number_of_accounts,omitempty"`
	AccountsOverflowTrimQuantity *uint64       `ic:"accounts_overflow_trim_quantity,omitempty" json:"accounts_overflow_trim_quantity,omitempty"`
	ArchiveOptions               struct {
		NumBlocksToArchive         uint64                 `ic:"num_blocks_to_archive" json:"num_blocks_to_archive"`
		MaxTransactionsPerResponse *uint64                `ic:"max_transactions_per_response,omitempty" json:"max_transactions_per_response,omitempty"`
		TriggerThreshold           uint64                 `ic:"trigger_threshold" json:"trigger_threshold"`
		MaxMessageSizeBytes        *uint64                `ic:"max_message_size_bytes,omitempty" json:"max_message_size_bytes,omitempty"`
		CyclesForArchiveCreation   *uint64                `ic:"cycles_for_archive_creation,omitempty" json:"cycles_for_archive_creation,omitempty"`
		NodeMaxMemorySizeBytes     *uint64                `ic:"node_max_memory_size_bytes,omitempty" json:"node_max_memory_size_bytes,omitempty"`
		ControllerId               principal.Principal    `ic:"controller_id" json:"controller_id"`
		MoreControllerIds          *[]principal.Principal `ic:"more_controller_ids,omitempty" json:"more_controller_ids,omitempty"`
	} `ic:"archive_options" json:"archive_options"`
}

type LedgerArg struct {
	Init    *InitArgs     `ic:"Init,variant"`
	Upgrade **UpgradeArgs `ic:"Upgrade,variant"`
}

type Map = []struct {
	Field0 string `ic:"0" json:"0"`
	Field1 Value  `ic:"1" json:"1"`
}

type MetadataValue struct {
	Nat  *idl.Nat `ic:"Nat,variant"`
	Int  *idl.Int `ic:"Int,variant"`
	Text *string  `ic:"Text,variant"`
	Blob *[]byte  `ic:"Blob,variant"`
}

type Mint struct {
	To            Account    `ic:"to" json:"to"`
	Memo          *[]byte    `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *Timestamp `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat    `ic:"amount" json:"amount"`
}

type QueryArchiveFn struct { /* NOT SUPPORTED */
}

type QueryBlockArchiveFn struct { /* NOT SUPPORTED */
}

type StandardRecord struct {
	Url  string `ic:"url" json:"url"`
	Name string `ic:"name" json:"name"`
}

type Subaccount = []byte

type Timestamp = uint64

type Tokens = idl.Nat

type Transaction struct {
	Burn      *Burn     `ic:"burn,omitempty" json:"burn,omitempty"`
	Kind      string    `ic:"kind" json:"kind"`
	Mint      *Mint     `ic:"mint,omitempty" json:"mint,omitempty"`
	Approve   *Approve  `ic:"approve,omitempty" json:"approve,omitempty"`
	Timestamp Timestamp `ic:"timestamp" json:"timestamp"`
	Transfer  *Transfer `ic:"transfer,omitempty" json:"transfer,omitempty"`
}

type TransactionRange struct {
	Transactions []Transaction `ic:"transactions" json:"transactions"`
}

type Transfer struct {
	To            Account    `ic:"to" json:"to"`
	Fee           *idl.Nat   `ic:"fee,omitempty" json:"fee,omitempty"`
	From          Account    `ic:"from" json:"from"`
	Memo          *[]byte    `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *Timestamp `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat    `ic:"amount" json:"amount"`
	Spender       *Account   `ic:"spender,omitempty" json:"spender,omitempty"`
}

type TransferArg struct {
	FromSubaccount *Subaccount `ic:"from_subaccount,omitempty" json:"from_subaccount,omitempty"`
	To             Account     `ic:"to" json:"to"`
	Amount         Tokens      `ic:"amount" json:"amount"`
	Fee            *Tokens     `ic:"fee,omitempty" json:"fee,omitempty"`
	Memo           *[]byte     `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime  *Timestamp  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
}

type TransferError struct {
	BadFee *struct {
		ExpectedFee Tokens `ic:"expected_fee" json:"expected_fee"`
	} `ic:"BadFee,variant"`
	BadBurn *struct {
		MinBurnAmount Tokens `ic:"min_burn_amount" json:"min_burn_amount"`
	} `ic:"BadBurn,variant"`
	InsufficientFunds *struct {
		Balance Tokens `ic:"balance" json:"balance"`
	} `ic:"InsufficientFunds,variant"`
	TooOld          *idl.Null `ic:"TooOld,variant"`
	CreatedInFuture *struct {
		LedgerTime Timestamp `ic:"ledger_time" json:"ledger_time"`
	} `ic:"CreatedInFuture,variant"`
	TemporarilyUnavailable *idl.Null `ic:"TemporarilyUnavailable,variant"`
	Duplicate              *struct {
		DuplicateOf BlockIndex `ic:"duplicate_of" json:"duplicate_of"`
	} `ic:"Duplicate,variant"`
	GenericError *struct {
		ErrorCode idl.Nat `ic:"error_code" json:"error_code"`
		Message   string  `ic:"message" json:"message"`
	} `ic:"GenericError,variant"`
}

type TransferFromArgs struct {
	SpenderSubaccount *Subaccount `ic:"spender_subaccount,omitempty" json:"spender_subaccount,omitempty"`
	From              Account     `ic:"from" json:"from"`
	To                Account     `ic:"to" json:"to"`
	Amount            Tokens      `ic:"amount" json:"amount"`
	Fee               *Tokens     `ic:"fee,omitempty" json:"fee,omitempty"`
	Memo              *[]byte     `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime     *Timestamp  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
}

type TransferFromError struct {
	BadFee *struct {
		ExpectedFee Tokens `ic:"expected_fee" json:"expected_fee"`
	} `ic:"BadFee,variant"`
	BadBurn *struct {
		MinBurnAmount Tokens `ic:"min_burn_amount" json:"min_burn_amount"`
	} `ic:"BadBurn,variant"`
	InsufficientFunds *struct {
		Balance Tokens `ic:"balance" json:"balance"`
	} `ic:"InsufficientFunds,variant"`
	InsufficientAllowance *struct {
		Allowance Tokens `ic:"allowance" json:"allowance"`
	} `ic:"InsufficientAllowance,variant"`
	TooOld          *idl.Null `ic:"TooOld,variant"`
	CreatedInFuture *struct {
		LedgerTime Timestamp `ic:"ledger_time" json:"ledger_time"`
	} `ic:"CreatedInFuture,variant"`
	Duplicate *struct {
		DuplicateOf BlockIndex `ic:"duplicate_of" json:"duplicate_of"`
	} `ic:"Duplicate,variant"`
	TemporarilyUnavailable *idl.Null `ic:"TemporarilyUnavailable,variant"`
	GenericError           *struct {
		ErrorCode idl.Nat `ic:"error_code" json:"error_code"`
		Message   string  `ic:"message" json:"message"`
	} `ic:"GenericError,variant"`
}

type TransferFromResult struct {
	Ok  *BlockIndex        `ic:"Ok,variant"`
	Err *TransferFromError `ic:"Err,variant"`
}

type TransferResult struct {
	Ok  *BlockIndex    `ic:"Ok,variant"`
	Err *TransferError `ic:"Err,variant"`
}

type TxIndex = idl.Nat

type UpgradeArgs struct {
	Metadata *[]struct {
		Field0 string        `ic:"0" json:"0"`
		Field1 MetadataValue `ic:"1" json:"1"`
	} `ic:"metadata,omitempty" json:"metadata,omitempty"`
	TokenSymbol                  *string               `ic:"token_symbol,omitempty" json:"token_symbol,omitempty"`
	TokenName                    *string               `ic:"token_name,omitempty" json:"token_name,omitempty"`
	TransferFee                  *idl.Nat              `ic:"transfer_fee,omitempty" json:"transfer_fee,omitempty"`
	ChangeFeeCollector           *ChangeFeeCollector   `ic:"change_fee_collector,omitempty" json:"change_fee_collector,omitempty"`
	MaxMemoLength                *uint16               `ic:"max_memo_length,omitempty" json:"max_memo_length,omitempty"`
	FeatureFlags                 *FeatureFlags         `ic:"feature_flags,omitempty" json:"feature_flags,omitempty"`
	MaximumNumberOfAccounts      *uint64               `ic:"maximum_number_of_accounts,omitempty" json:"maximum_number_of_accounts,omitempty"`
	AccountsOverflowTrimQuantity *uint64               `ic:"accounts_overflow_trim_quantity,omitempty" json:"accounts_overflow_trim_quantity,omitempty"`
	ChangeArchiveOptions         *ChangeArchiveOptions `ic:"change_archive_options,omitempty" json:"change_archive_options,omitempty"`
}

type Value struct {
	Blob  *[]byte  `ic:"Blob,variant"`
	Text  *string  `ic:"Text,variant"`
	Nat   *idl.Nat `ic:"Nat,variant"`
	Nat64 *uint64  `ic:"Nat64,variant"`
	Int   *idl.Int `ic:"Int,variant"`
	Array *[]Value `ic:"Array,variant"`
	Map   *Map     `ic:"Map,variant"`
}
