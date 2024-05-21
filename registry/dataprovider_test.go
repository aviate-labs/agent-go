package registry

import (
	"testing"
)

func TestDataProvider_GetChangesSince(t *testing.T) {
	dp, err := NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = dp.GetChangesSince(0); err != nil {
		t.Fatal(err)
	}
}

func TestDataProvider_GetLatestVersion(t *testing.T) {
	dp, err := NewDataProvider()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = dp.GetLatestVersion(); err != nil {
		t.Fatal(err)
	}
}
