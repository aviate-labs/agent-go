package pocketic_test

import (
	"bytes"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/pocketic"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var (
	s, setupErr = pocketic.New(pocketic.DefaultSubnetConfig)
	wasmModule  = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
)

func TestPocketIC(t *testing.T) {
	dfxPath, err := exec.LookPath("dfx")
	if err != nil {
		t.Skip(err)
	}
	var out bytes.Buffer
	dfxCacheCmd := exec.Command(dfxPath, "cache", "show")
	dfxCacheCmd.Stdout = &out
	if err := dfxCacheCmd.Run(); err != nil {
		t.Skip(err)
	}
	mocPath := fmt.Sprintf("%s/moc", strings.TrimSpace(out.String()))
	mocCmd := exec.Command(mocPath, "testdata/main.mo", "-o", "testdata/main.wasm", "--idl")
	if err := mocCmd.Run(); err != nil {
		t.Skip(err)
	}
	helloWasm, err := os.Open("testdata/main.wasm")
	if err != nil {
		t.Skip(err)
	}
	canisterID, err := s.CreateAndInstallCanister(helloWasm, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var helloWorld string
	if err := s.QueryCall(*canisterID, "helloQuery", []any{"world"}, []any{&helloWorld}); err != nil {
		t.Fatal(err)
	}
	if helloWorld != "Hello, world!" {
		t.Errorf("hello world is %s, expected Hello, world!", helloWorld)
	}
	if err := s.UpdateCall(*canisterID, "helloUpdate", []any{"there"}, []any{&helloWorld}); err != nil {
		t.Fatal(err)
	}
	if helloWorld != "Hello, there!" {
		t.Errorf("hello world is %s, expected Hello, there!", helloWorld)
	}

	host, err := url.Parse(s.GetHost())
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewAgent(*canisterID, agent.Config{
		FetchRootKey: true,
		ClientConfig: &agent.ClientConfig{
			Host: host,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Agent QueryCall", func(t *testing.T) {
		resp, err := a.HelloQuery("world")
		if err != nil {
			t.Fatal(err)
		}
		if *resp != "Hello, world!" {
			t.Errorf("resp is %s, expected Hello, world!", *resp)
		}
	})
	t.Run("Agent UpdateCall", func(t *testing.T) {
		t.Skip("PocketIC does not advance automatically, so the time is not updated.")

		resp, err := a.HelloUpdate("there")
		if err != nil {
			t.Fatal(err)
		}
		if *resp != "Hello, there!" {
			t.Errorf("resp is %s, expected Hello, there!", *resp)
		}
	})
}

func TestPocketIC_CreateAndInstallCanister(t *testing.T) {
	if _, err := s.CreateAndInstallCanister(bytes.NewReader(wasmModule), nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPocketIC_CreateCanister(t *testing.T) {
	cID, err := principal.Decode("rwlgt-iiaaa-aaaaa-aaaaa-cai")
	if err != nil {
		t.Fatal(err)
	}
	canisterID, err := s.CreateCanister(pocketic.CreateCanisterArgs{
		SpecifiedID: &cID,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if canisterID == nil {
		t.Fatal("canister ID is nil")
	}
	if canisterID.String() != "rwlgt-iiaaa-aaaaa-aaaaa-cai" {
		t.Errorf("canister ID is %s, expected rwlgt-iiaaa-aaaaa-aaaaa-cai", canisterID.String())
	}

	if _, err := s.GetSubnet(cID); err != nil {
		t.Fatal(err)
	}

	amount := 1_000_000_000_000
	cyclesBefore, err := s.GetCycleBalance(cID)
	if err != nil {
		t.Fatal(err)
	}
	cyclesAfter, err := s.AddCycles(cID, amount)
	if err != nil {
		t.Fatal(err)
	}
	if cyclesAfter-cyclesBefore != amount {
		t.Errorf("cycles added is %d, expected %d", cyclesAfter-cyclesBefore, amount)
	}

	if _, err := s.CreateCanister(pocketic.CreateCanisterArgs{
		SpecifiedID: canisterID,
	}, nil); err == nil {
		t.Error("expected error")
	}
	if err := s.InstallCode(cID, bytes.NewReader(wasmModule), nil); err != nil {
		t.Fatal(err)
	}
}

func TestPocketIC_GetRootKey(t *testing.T) {
	rootKey, err := s.GetRootKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(rootKey) == 0 {
		t.Error("root key is empty")
	}
}

func TestPocketIC_Time(t *testing.T) {
	t.Run("GetTime", func(t *testing.T) {
		ns, err := s.GetTime()
		if err != nil {
			t.Fatal(err)
		}
		if ns.IsZero() {
			t.Error("time is zero")
		}
	})

	t.Run("SetTime", func(t *testing.T) {
		n := time.Now().Nanosecond()
		if err := s.SetTime(n); err != nil {
			t.Fatal(err)
		}
		ns, err := s.GetTime()
		if err != nil {
			t.Fatal(err)
		}
		if ns.Nanosecond() != n {
			t.Errorf("time is %d, expected %d", ns.Nanosecond(), n)
		}
	})

	t.Run("AdvanceTime", func(t *testing.T) {
		ns, err := s.GetTime()
		if err != nil {
			t.Fatal(err)
		}
		if err := s.AdvanceTime(10); err != nil {
			t.Fatal(err)
		}
		ns2, err := s.GetTime()
		if err != nil {
			t.Fatal(err)
		}
		if !ns2.After(*ns) {
			t.Errorf("time is %v, expected after %v", ns2, ns)
		}
	})

	t.Run("Tick", func(t *testing.T) {
		if err := s.Tick(); err != nil {
			t.Fatal(err)
		}
	})
}

func init() {
	if setupErr != nil {
		panic(setupErr)
	}
}
