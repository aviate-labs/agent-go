package pocketic

import (
	"encoding/json"
	"testing"
)

func TestNone(t *testing.T) {
	var none None
	raw, err := json.Marshal(none)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "\"None\"" {
		t.Fatalf("expected None, got %s", string(raw))
	}
	var decoded None
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
}
