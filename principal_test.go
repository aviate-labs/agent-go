package agent_test

import (
	"encoding/hex"
	"fmt"

	"github.com/allusion-be/agent-go"
)

func ExamplePrincipal() {
	raw, _ := hex.DecodeString("abcd01")
	p := agent.Principal(raw)
	fmt.Println(p.Encode())
	// Output:
	// em77e-bvlzu-aq
}

func ExampleDecodePrincipal() {
	p, _ := agent.DecodePrincipal("em77e-bvlzu-aq")
	fmt.Printf("%x", p)
	// Output:
	// abcd01
}
