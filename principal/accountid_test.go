package principal_test

import (
	"encoding/json"
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
	for i := range 100 {
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

func TestAccountIdentifier_MarshalJSON(t *testing.T) {
	original := principal.NewAccountID(principal.AnonymousID, principal.DefaultSubAccount)
	raw, err := json.Marshal(original)
	if err != nil {
		t.Error(err)
	}
	var decoded principal.AccountIdentifier
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Error(err)
	}
	if original != decoded {
		t.Errorf("expected %v, got %v", original, decoded)
	}
}
