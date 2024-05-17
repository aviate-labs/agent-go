package registry

import (
	"testing"
)

func TestDataProvider_GetChangesSince(t *testing.T) {
	if _, _, err := new(DataProvider).GetChangesSince(0); err != nil {
		t.Fatal(err)
	}
}
