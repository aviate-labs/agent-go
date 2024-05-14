// Package index provides a client for the "index" canister.
// Do NOT edit this file. It was automatically generated by https://github.com/aviate-labs/agent-go.
package index

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/principal"
)

type Account struct {
	Owner      principal.Principal `ic:"owner" json:"owner"`
	Subaccount *[]byte             `ic:"subaccount,omitempty" json:"subaccount,omitempty"`
}

// Agent is a client for the "index" canister.
type Agent struct {
	*agent.Agent
	CanisterId principal.Principal
}

// NewAgent creates a new agent for the "index" canister.
func NewAgent(canisterId principal.Principal, config agent.Config) (*Agent, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}
	return &Agent{
		Agent:      a,
		CanisterId: canisterId,
	}, nil
}

// GetAccountTransactions calls the "get_account_transactions" method on the "index" canister.
func (a Agent) GetAccountTransactions(arg0 GetAccountTransactionsArgs) (*GetTransactionsResult, error) {
	var r0 GetTransactionsResult
	if err := a.Agent.Call(
		a.CanisterId,
		"get_account_transactions",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// LedgerId calls the "ledger_id" method on the "index" canister.
func (a Agent) LedgerId() (*principal.Principal, error) {
	var r0 principal.Principal
	if err := a.Agent.Query(
		a.CanisterId,
		"ledger_id",
		[]any{},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

// ListSubaccounts calls the "list_subaccounts" method on the "index" canister.
func (a Agent) ListSubaccounts(arg0 ListSubaccountsArgs) (*[]SubAccount, error) {
	var r0 []SubAccount
	if err := a.Agent.Query(
		a.CanisterId,
		"list_subaccounts",
		[]any{arg0},
		[]any{&r0},
	); err != nil {
		return nil, err
	}
	return &r0, nil
}

type Approve struct {
	Fee               *idl.Nat `ic:"fee,omitempty" json:"fee,omitempty"`
	From              Account  `ic:"from" json:"from"`
	Memo              *[]uint8 `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime     *uint64  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount            idl.Nat  `ic:"amount" json:"amount"`
	ExpectedAllowance *idl.Nat `ic:"expected_allowance,omitempty" json:"expected_allowance,omitempty"`
	ExpiresAt         *uint64  `ic:"expires_at,omitempty" json:"expires_at,omitempty"`
	Spender           Account  `ic:"spender" json:"spender"`
}

type Burn struct {
	From          Account  `ic:"from" json:"from"`
	Memo          *[]uint8 `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *uint64  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat  `ic:"amount" json:"amount"`
	Spender       *Account `ic:"spender,omitempty" json:"spender,omitempty"`
}

type GetAccountTransactionsArgs struct {
	Account    Account `ic:"account" json:"account"`
	Start      *TxId   `ic:"start,omitempty" json:"start,omitempty"`
	MaxResults idl.Nat `ic:"max_results" json:"max_results"`
}

type GetTransactions struct {
	Transactions []TransactionWithId `ic:"transactions" json:"transactions"`
	OldestTxId   *TxId               `ic:"oldest_tx_id,omitempty" json:"oldest_tx_id,omitempty"`
}

type GetTransactionsErr struct {
	Message string `ic:"message" json:"message"`
}

type GetTransactionsResult struct {
	Ok  *GetTransactions    `ic:"Ok,variant"`
	Err *GetTransactionsErr `ic:"Err,variant"`
}

type InitArgs struct {
	LedgerId principal.Principal `ic:"ledger_id" json:"ledger_id"`
}

type ListSubaccountsArgs struct {
	Owner principal.Principal `ic:"owner" json:"owner"`
	Start *SubAccount         `ic:"start,omitempty" json:"start,omitempty"`
}

type Mint struct {
	To            Account  `ic:"to" json:"to"`
	Memo          *[]uint8 `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *uint64  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat  `ic:"amount" json:"amount"`
}

type SubAccount = []byte

type Transaction struct {
	Burn      *Burn     `ic:"burn,omitempty" json:"burn,omitempty"`
	Kind      string    `ic:"kind" json:"kind"`
	Mint      *Mint     `ic:"mint,omitempty" json:"mint,omitempty"`
	Approve   *Approve  `ic:"approve,omitempty" json:"approve,omitempty"`
	Timestamp uint64    `ic:"timestamp" json:"timestamp"`
	Transfer  *Transfer `ic:"transfer,omitempty" json:"transfer,omitempty"`
}

type TransactionWithId struct {
	Id          TxId        `ic:"id" json:"id"`
	Transaction Transaction `ic:"transaction" json:"transaction"`
}

type Transfer struct {
	To            Account  `ic:"to" json:"to"`
	Fee           *idl.Nat `ic:"fee,omitempty" json:"fee,omitempty"`
	From          Account  `ic:"from" json:"from"`
	Memo          *[]uint8 `ic:"memo,omitempty" json:"memo,omitempty"`
	CreatedAtTime *uint64  `ic:"created_at_time,omitempty" json:"created_at_time,omitempty"`
	Amount        idl.Nat  `ic:"amount" json:"amount"`
	Spender       *Account `ic:"spender,omitempty" json:"spender,omitempty"`
}

type TxId = idl.Nat
