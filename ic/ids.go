package ic

import "github.com/aviate-labs/agent-go/principal"

var (
	// https://internetcomputer.org/docs/current/references/ic-interface-spec#ic-management-canister
	MANAGEMENT_CANISTER_PRINCIPAL, _ = principal.Decode("aaaaa-aa")
)
