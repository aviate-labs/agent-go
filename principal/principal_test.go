package principal_test

import (
	"encoding/hex"
	"fmt"

	"github.com/aviate-labs/agent-go/principal"
)

func ExampleDecode() {
	p, _ := principal.Decode("em77e-bvlzu-aq")
	fmt.Printf("%x", p.Raw)
	// Output:
	// abcd01
}

func ExamplePrincipal() {
	raw, _ := hex.DecodeString("abcd01")
	p := principal.Principal{raw}
	fmt.Println(p.Encode())
	// Output:
	// em77e-bvlzu-aq
}
