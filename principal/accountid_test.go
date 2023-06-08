package principal_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
)

func ExampleFromPrincipal() {
	fmt.Println(principal.FromPrincipal(principal.Principal{}, principal.DefaultSubAccount).String())
	// Output:
	// 2d0e897f7e862d2b57d9bc9ea5c65f9a24ac6c074575f47898314b8d6cb0929d
}
