package ledger

import (
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/candid/marshal"
	"github.com/aviate-labs/agent-go/principal"
)

func accountIdToVec(accountId principal.AccountIdentifier) []any {
	var vec []any
	for _, v := range accountId.Bytes() {
		vec = append(vec, v)
	}
	return vec
}

func subAccountIdToVec(accountId principal.SubAccount) []any {
	var vec []any
	for _, v := range accountId[:] {
		vec = append(vec, v)
	}
	return vec
}

// AccountBalanceArgs is an argument to the `account_balance` method.
type AccountBalanceArgs struct {
	Account AccountIdentifier
}

func (args AccountBalanceArgs) encode() ([]byte, error) {
	m := make(map[string]any)
	m["account"] = accountIdToVec(args.Account)
	return marshal.Marshal([]any{m})
}

// AccountIdentifier is a 32-byte array.
// The first 4 bytes is big-endian encoding of a CRC32 checksum of the last 28 bytes.
type AccountIdentifier = principal.AccountIdentifier

// BlockIndex the sequence number of a block produced by the ledger.
type BlockIndex = uint64

// Memo an arbitrary number associated with a transaction.
// The caller can set it in a `transfer` call as a correlation identifier.
type Memo = uint64

// SubAccount is an arbitrary 32-byte byte array.
// Ledger uses sub-accounts to compute the source address, which enables one
// principal to control multiple ledger accounts.
type SubAccount = principal.SubAccount

// TimeStamp is the number of nanoseconds from the UNIX epoch in UTC timezone.
type TimeStamp struct {
	TimestampNanos uint64
}

func (t TimeStamp) toMap() map[string]any {
	m := make(map[string]any)
	m["timestamp_nanos"] = t.TimestampNanos
	return m
}

// Tokens is the amount of tokens, measured in 10^-8 of a token.
type Tokens struct {
	E8S uint64
}

func recordTokens(m map[string]any) (*Tokens, bool) {
	if e8s, ok := m[idl.HashString("e8s")]; ok {
		return &Tokens{
			E8S: e8s.(uint64),
		}, true
	}
	return nil, false
}

func (t Tokens) toMap() map[string]any {
	m := make(map[string]any)
	m["e8s"] = t.E8S
	return m
}

// TransferArgs is an argument to the `transfer` method.
type TransferArgs struct {
	// Transaction memo.
	// See comments for the `Memo` type.
	Memo Memo
	// The amount that the caller wants to transfer to the destination address.
	Amount Tokens
	// The amount that the caller pays for the transaction.
	// Must be 10000 e8s.
	Fee Tokens
	// The sub-account from which the caller wants to transfer funds.
	// If null, the ledger uses the default (all zeros) subaccount to compute the source address.
	// See comments for the `SubAccount` type.
	FromSubAccount *SubAccount
	// The destination account.
	// If the transfer is successful, the balance of this address increases by `amount`.
	To AccountIdentifier
	// The point in time when the caller created this request.
	// If null, the ledger uses current IC time as the timestamp.
	CreatedAtTime *TimeStamp
}

func (args TransferArgs) encode() ([]byte, error) {
	m := make(map[string]any)
	m["memo"] = args.Memo
	m["amount"] = args.Amount.toMap()
	m["fee"] = args.Fee.toMap()
	if args.FromSubAccount != nil {
		m["from_subaccount"] = subAccountIdToVec(*args.FromSubAccount)
	}
	m["to"] = accountIdToVec(args.To)
	if args.CreatedAtTime != nil {
		m["created_at_time"] = args.CreatedAtTime.toMap()
	}
	return marshal.Marshal([]any{m})
}

// TransferError is an error returned by the `transfer` method.
type TransferError struct {
	// The fee that the caller specified in the transfer request was not the one that ledger expects.
	// The caller can change the transfer fee to the `expected_fee` and retry the request.
	BadFee *struct{ ExpectedFee Tokens }
	// The account specified by the caller doesn't have enough funds.
	InsufficientFunds *struct{ Balance Tokens }
	// The request is too old.
	// The ledger only accepts requests created within 24 hours window.
	// This is a non-recoverable error.
	TxTooOld *struct{ AllowedWindowNanos uint64 }
	// The caller specified `created_at_time` that is too far in future.
	// The caller can retry the request later.
	TxCreatedInFuture *struct{}
	// The ledger has already executed the request.
	// `duplicate_of` field is equal to the index of the block containing the original transaction.
	TxDuplicate *struct{ DuplicateOf BlockIndex }
}

// TransferResult is a result of the `transfer` method.
type TransferResult struct {
	Ok  *BlockIndex
	Err *TransferError
}
