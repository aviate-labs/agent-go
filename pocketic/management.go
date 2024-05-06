package pocketic

import (
	"fmt"
	"net/http"

	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/ic"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/principal"
)

// AddCycles add cycles to a canister. Returns the new balance.
func (pic PocketIC) AddCycles(canisterID principal.Principal, amount int) (int, error) {
	var resp struct {
		Cycles int `json:"cycles"`
	}
	if err := pic.do(
		http.MethodPost,
		fmt.Sprintf("%s/update/add_cycles", pic.InstanceURL()),
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
	return pic.CreateCanisterWithArgs(ProvisionalCreateCanisterArgument{})
}

// CreateCanisterOnSubnet creates a canister on the specified subnet with the specified settings.
func (pic PocketIC) CreateCanisterOnSubnet(subnetID principal.Principal, args ProvisionalCreateCanisterArgument) (*principal.Principal, error) {
	return pic.createCanister(&RawEffectivePrincipalSubnetID{SubnetID: subnetID.Raw}, args)
}

// CreateCanisterWithArgs creates a canister with the specified settings and cycles.
func (pic PocketIC) CreateCanisterWithArgs(args ProvisionalCreateCanisterArgument) (*principal.Principal, error) {
	return pic.createCanister(new(RawEffectivePrincipalNone), args)
}

// CreateCanisterWithID creates a canister with the specified canister ID and settings.
func (pic PocketIC) CreateCanisterWithID(canisterID principal.Principal, args ProvisionalCreateCanisterArgument) (*principal.Principal, error) {
	return pic.createCanister(&RawEffectivePrincipalCanisterID{CanisterID: canisterID.Raw}, args)
}

// InstallCode installs a canister with the specified wasm module and arguments.
func (pic PocketIC) InstallCode(canisterID principal.Principal, wasmModule []byte, arg []byte, optSender *principal.Principal) error {
	sender := principal.AnonymousID
	if optSender != nil {
		sender = *optSender
	}
	payload, err := idl.Marshal([]any{ic0.InstallCodeArgs{
		Mode: ic0.CanisterInstallMode{
			Install: new(idl.Null),
		},
		WasmModule: wasmModule,
		CanisterId: canisterID,
		Arg:        arg,
	}})
	if err != nil {
		return err
	}
	_, err = pic.updateCallWithEP(
		ic.MANAGEMENT_CANISTER_PRINCIPAL,
		&RawEffectivePrincipalCanisterID{CanisterID: canisterID.Raw},
		sender,
		"install_code",
		payload,
	)
	return err
}

// UninstallCode uninstalls a canister.
func (pic PocketIC) UninstallCode(canisterID principal.Principal, optSender *principal.Principal) error {
	sender := principal.AnonymousID
	if optSender != nil {
		sender = *optSender
	}
	payload, err := idl.Marshal([]any{ic0.UninstallCodeArgs{
		CanisterId: canisterID,
	}})
	if err != nil {
		return err
	}
	_, err = pic.updateCallWithEP(
		ic.MANAGEMENT_CANISTER_PRINCIPAL,
		&RawEffectivePrincipalCanisterID{CanisterID: canisterID.Raw},
		sender,
		"uninstall_code",
		payload,
	)
	return err
}

func (pic PocketIC) createCanister(effectivePrincipal RawEffectivePrincipal, args ProvisionalCreateCanisterArgument) (*principal.Principal, error) {
	payload, err := idl.Marshal([]any{args})
	if err != nil {
		return nil, err
	}
	raw, err := pic.updateCallWithEP(
		ic.MANAGEMENT_CANISTER_PRINCIPAL,
		effectivePrincipal,
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
