package agent_test

import (
	"fmt"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/candid-go"
	"github.com/aviate-labs/principal-go"
)

func ExampleQuery() {
	ledgerID, _ := principal.Decode("ryjl3-tyaaa-aaaaa-aaaba-cai")
	agent := agent.New()
	args, _ := candid.EncodeValue("record { account = \"609d3e1e45103a82adc97d4f88c51f78dedb25701e8e51e8c4fec53448aadc29\" }")
	_, _, err := agent.Query(ledgerID, "account_balance_dfx", args)
	fmt.Println(err)
	// Output:
	// <nil>
}
