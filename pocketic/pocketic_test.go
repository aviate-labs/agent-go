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
	"sync"
	"testing"
	"time"
)

func TestConcurrentCalls(t *testing.T) {
	pic, err := pocketic.New(pocketic.WithPollingDelay(10*time.Millisecond, 10*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()

			canisterID, err := pic.CreateCanister()
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := pic.AddCycles(*canisterID, 2_000_000_000_000); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestCreateCanister(t *testing.T) {
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
}

func TestHttpGateway(t *testing.T) {
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

	agentConfig := agent.Config{
		ClientConfig: &agent.ClientConfig{Host: host},
		FetchRootKey: true,
		Logger:       new(testLogger),
	}
	mgmtAgent, err := ic0.NewAgent(ic.MANAGEMENT_CANISTER_PRINCIPAL, agentConfig)
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

	wasmModule := compileMotoko(t, "testdata/main.mo", "testdata/main.wasm")
	if err := mgmtAgent.InstallCode(ic0.InstallCodeArgs{
		Mode: ic0.CanisterInstallMode{
			Install: new(idl.Null),
		},
		CanisterId: result.CanisterId,
		WasmModule: wasmModule,
	}); err != nil {
		t.Fatal(err)
	}

	helloAgent, err := NewAgent(result.CanisterId, agentConfig)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := helloAgent.HelloUpdate("world")
	if err != nil {
		t.Fatal(err)
	}
	if *resp != "Hello, world!" {
		t.Fatalf("unexpected response: %s", *resp)
	}

	if err := pic.MakeDeterministic(); err != nil {
		t.Fatal(err)
	}
}

func compileMotoko(t *testing.T, in, out string) []byte {
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
	wasmModule, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return wasmModule
}

type testLogger struct{}

func (t testLogger) Printf(format string, v ...any) {
	fmt.Printf(format+"\n", v...)
}
