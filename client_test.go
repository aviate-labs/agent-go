package agent_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/niccolofant/agent-go"
	"github.com/niccolofant/agent-go/principal"
)

func recordingClient(t *testing.T, opts ...agent.ClientOption) (agent.Client, *string) {
	t.Helper()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		// Minimal replied-call response; ReadState returns the body verbatim.
		body, _ := cbor.Marshal(map[string]any{"status": "replied", "certificate": []byte{}})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	host, _ := url.Parse(srv.URL)
	c := agent.NewClient(append([]agent.ClientOption{agent.WithHostURL(host)}, opts...)...)
	return c, &gotPath
}

func TestClientCallDefaultsToV4(t *testing.T) {
	c, gotPath := recordingClient(t)
	cid := principal.MustDecode("aaaaa-aa")
	if _, err := c.Call(context.Background(), cid, nil); err != nil {
		t.Fatal(err)
	}
	want := "/api/v4/canister/" + cid.Encode() + "/call"
	if *gotPath != want {
		t.Fatalf("got %q, want %q", *gotPath, want)
	}
}

func TestClientReadStateDefaultsToV3(t *testing.T) {
	c, gotPath := recordingClient(t)
	cid := principal.MustDecode("aaaaa-aa")
	if _, err := c.ReadState(context.Background(), cid, nil); err != nil {
		t.Fatal(err)
	}
	want := "/api/v3/canister/" + cid.Encode() + "/read_state"
	if *gotPath != want {
		t.Fatalf("got %q, want %q", *gotPath, want)
	}
}

func TestClientLegacyAPI(t *testing.T) {
	cid := principal.MustDecode("aaaaa-aa")

	cCall, callPath := recordingClient(t, agent.WithLegacyAPI())
	if _, err := cCall.Call(context.Background(), cid, nil); err != nil {
		t.Fatal(err)
	}
	if want := "/api/v3/canister/" + cid.Encode() + "/call"; *callPath != want {
		t.Fatalf("call: got %q, want %q", *callPath, want)
	}

	cRead, readPath := recordingClient(t, agent.WithLegacyAPI())
	if _, err := cRead.ReadState(context.Background(), cid, nil); err != nil {
		t.Fatal(err)
	}
	if want := "/api/v2/canister/" + cid.Encode() + "/read_state"; *readPath != want {
		t.Fatalf("read_state: got %q, want %q", *readPath, want)
	}
}
