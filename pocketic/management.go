package pocketic

import (
	"fmt"
	"net/http"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/ic"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/principal"
)

func (pic PocketIC) AddCycles(canisterID principal.Principal, amount int) (int, error) {
	var resp struct {
		Cycles int `json:"cycles"`
	}
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/add_cycles", pic.instanceURL()),
		http.StatusOK,
		RawAddCycles{
			Amount:     amount,
			CanisterID: canisterID.Raw,
		},
		&resp,
	); err != nil {
		return 0, err
	}
	return resp.Cycles, nil
}

// CreateCanister creates a canister with default settings as the anonymous principal.
func (pic PocketIC) CreateCanister() (*principal.Principal, error) {
	payload, err := idl.Marshal([]any{ProvisionalCreateCanisterArgument{}})
	if err != nil {
		return nil, err
	}
	raw, err := pic.updateCallWithEP(
		ic.MANAGEMENT_CANISTER_PRINCIPAL,
		new(RawEffectivePrincipalNone),
		principal.AnonymousID,
		"provisional_create_canister_with_cycles",
		payload,
	)
	if err != nil {
		return nil, err
	}
	var resp struct {
		CanisterID principal.Principal `ic:"canister_id"`
	}
	if err := idl.Unmarshal(raw, []any{&resp}); err != nil {
		return nil, err
	}
	return &resp.CanisterID, nil
}

type ProvisionalCreateCanisterArgument struct {
	Settings    *ic0.CanisterSettings `ic:"settings"`
	SpecifiedID *principal.Principal  `ic:"specified_id"`
	Amount      *int                  `ic:"amount"`
}
