package ic

import "github.com/aviate-labs/agent-go/principal"

var (
	// https://dashboard.internetcomputer.org/canister/rwlgt-iiaaa-aaaaa-aaaaa-cai
	REGISTRY_PRINCIPAL, _ = principal.Decode("rwlgt-iiaaa-aaaaa-aaaaa-cai")
	// https://dashboard.internetcomputer.org/canister/rrkah-fqaaa-aaaaa-aaaaq-cai
	GOVERNANCE_PRINCIPAL, _ = principal.Decode("rrkah-fqaaa-aaaaa-aaaaq-cai")
	// https://dashboard.internetcomputer.org/canister/ryjl3-tyaaa-aaaaa-aaaba-cai
	LEDGER_PRINCIPAL, _ = principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	// https://dashboard.internetcomputer.org/canister/r7inp-6aaaa-aaaaa-aaabq-cai
	NNS_ROOT_PRINCIPAL, _ = principal.Decode("r7inp-6aaaa-aaaaa-aaabq-cai")
	// https://dashboard.internetcomputer.org/canister/rkp4c-7iaaa-aaaaa-aaaca-cai
	CYCLES_MINTING_PRINCIPAL, _ = principal.Decode("rkp4c-7iaaa-aaaaa-aaaca-cai")
	// https://dashboard.internetcomputer.org/canister/rno2w-sqaaa-aaaaa-aaacq-cai
	LIFELINE_PRINCIPAL, _ = principal.Decode("rno2w-sqaaa-aaaaa-aaacq-cai")
	// https://dashboard.internetcomputer.org/canister/renrk-eyaaa-aaaaa-aaada-cai
	GENESIS_TOKEN_PRINCIPAL, _ = principal.Decode("renrk-eyaaa-aaaaa-aaada-cai")
	// https://dashboard.internetcomputer.org/canister/qaa6y-5yaaa-aaaaa-aaafa-cai
	SNS_WASM_PRINCIPAL, _ = principal.Decode("qaa6y-5yaaa-aaaaa-aaafa-cai")
	// https://dashboard.internetcomputer.org/canister/rdmx6-jaaaa-aaaaa-aaadq-cai
	IDENTITY_PRINCIPAL, _ = principal.Decode("rdmx6-jaaaa-aaaaa-aaadq-cai")
	// https://dashboard.internetcomputer.org/canister/qoctq-giaaa-aaaaa-aaaea-cai
	NNS_UI_PRINCIPAL, _ = principal.Decode("qoctq-giaaa-aaaaa-aaaea-cai")
)
