package icpledger_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/icpledger"
)

func ExampleAgent_Archives() {
	a, _ := icpledger.NewAgent(ic.LEDGER_PRINCIPAL, agent.Config{})
	archives, _ := a.Archives()
	fmt.Println(archives)
	// Output:
	// &{[{qjdve-lqaaa-aaaaa-aaaeq-cai}]}
}
