package pocketic_test

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/ic"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/pocketic"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func CreateCanister(t *testing.T) *pocketic.PocketIC {
	pic, err := pocketic.New(pocketic.WithLogger(new(testLogger)))
	if err != nil {
		t.Fatal(err)
	}

	canisterID, err := pic.CreateCanister()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pic.AddCycles(*canisterID, 2_000_000_000_000); err != nil {
		t.Fatal(err)
	}

	return pic
}

func MakeLive(t *testing.T) *pocketic.PocketIC {
	pic, err := pocketic.New(
		pocketic.WithLogger(new(testLogger)),
		pocketic.WithNNSSubnet(),
		pocketic.WithApplicationSubnet(),
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := pic.MakeLive(nil)
	if err != nil {
		t.Fatal(err)
	}
	host, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	mgmtAgent, err := ic0.NewAgent(ic.MANAGEMENT_CANISTER_PRINCIPAL, agent.Config{
		ClientConfig: &agent.ClientConfig{Host: host},
		FetchRootKey: true,
		Logger:       new(testLogger),
	})
	if err != nil {
		t.Fatal(err)
	}

	var ecID principal.Principal
	for _, t := range pic.Topology() {
		if t.SubnetKind == pocketic.ApplicationSubnet {
			ecID = t.CanisterRanges[0].Start
			break
		}
	}

	var result ic0.ProvisionalCreateCanisterWithCyclesResult
	createCall, err := mgmtAgent.ProvisionalCreateCanisterWithCyclesCall(ic0.ProvisionalCreateCanisterWithCyclesArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if err := createCall.WithEffectiveCanisterID(ecID).CallAndWait(&result); err != nil {
		t.Fatal(err)
	}

	compileMotoko(t, "testdata/main.mo", "testdata/main.wasm")
	wasmModule, err := os.ReadFile("testdata/main.wasm")
	if err != nil {
		t.Fatal(err)
	}
	if err := mgmtAgent.InstallCode(ic0.InstallCodeArgs{
		Mode: ic0.CanisterInstallMode{
			Install: new(idl.Null),
		},
		CanisterId: result.CanisterId,
		WasmModule: wasmModule,
	}); err != nil {
		t.Fatal(err)
	}

	return pic
}

func TestPocketIC(t *testing.T) {
	var instances []*pocketic.PocketIC
	t.Run("CreateCanister", func(t *testing.T) {
		instances = append(instances, CreateCanister(t))
	})
	t.Run("MakeLive", func(t *testing.T) {
		instances = append(instances, MakeLive(t))
	})
	for _, i := range instances {
		if err := i.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func compileMotoko(t *testing.T, in, out string) {
	dfxPath, err := exec.LookPath("dfx")
	if err != nil {
		t.Skipf("dfx not found: %v", err)
	}
	cmd := exec.Command(dfxPath, "cache", "show")
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	mocPath := path.Join(strings.TrimSpace(string(raw)), "moc")
	cmd = exec.Command(mocPath, in, "-o", out)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

type testLogger struct{}

func (t testLogger) Printf(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}
