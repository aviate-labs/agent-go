package pocketic

import (
	"bytes"
	"encoding/json"
	"github.com/aviate-labs/agent-go/ic"
	"testing"
)

func TestBase64EncodedBlob(t *testing.T) {
	blob := Base64EncodedBlob("Hello, there!")
	jsonEncoded, err := json.Marshal(blob)
	if err != nil {
		t.Fatal(err)
	}
	if string(jsonEncoded) != `"SGVsbG8sIHRoZXJlIQ=="` {
		t.Errorf("unexpected JSON encoding: %s", jsonEncoded)
	}
	var decoded Base64EncodedBlob
	if err := json.Unmarshal(jsonEncoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if string(decoded) != "Hello, there!" {
		t.Errorf("unexpected JSON decoding: %s", decoded)
	}
}

func TestRawCanisterID(t *testing.T) {
	canisterID := RawCanisterID{
		CanisterID: ic.LEDGER_PRINCIPAL.Raw,
	}
	jsonEncoded, err := json.Marshal(canisterID)
	if err != nil {
		t.Fatal(err)
	}
	if string(jsonEncoded) != `{"canister_id":"AAAAAAAAAAIBAQ=="}` {
		t.Errorf("unexpected JSON encoding: %s", jsonEncoded)
	}
	var decoded RawCanisterID
	if err := json.Unmarshal(jsonEncoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.CanisterID, ic.LEDGER_PRINCIPAL.Raw) {
		t.Errorf("unexpected JSON decoding: %s", decoded.CanisterID)
	}
}
