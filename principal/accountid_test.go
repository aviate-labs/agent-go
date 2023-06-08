package principal_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go/principal"
	"testing"
)

func ExampleNewAccountID() {
	fmt.Println(principal.NewAccountID(principal.Principal{}, principal.DefaultSubAccount).String())
	// Output:
	// 2d0e897f7e862d2b57d9bc9ea5c65f9a24ac6c074575f47898314b8d6cb0929d
}

func TestAccountIdentifier(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := principal.NewAccountID(principal.Principal{}, [32]byte{byte(i)})
		accountID, err := principal.DecodeAccountID(a.String())
		if err != nil {
			t.Error(err)
		}
		if accountID != a {
			t.Errorf("expected %v, got %v", a, accountID)
		}
	}
}
