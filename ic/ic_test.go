package ic_test

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/assetstorage"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/pocketic"
	"github.com/aviate-labs/agent-go/principal"
)

func TestModules(t *testing.T) {
	pic, err := pocketic.New()
	if err != nil {
		t.Skip(err)
	}

	rawHost, err := pic.MakeLive(nil)
	if err != nil {
		t.Fatal(err)
	}
	host, err := url.Parse(rawHost)
	if err != nil {
		t.Fatal(err)
	}

	config := agent.Config{
		ClientConfig: &agent.ClientConfig{Host: host},
		FetchRootKey: true,
		Logger:       new(localLogger),
	}

	t.Run("assetstorage", func(t *testing.T) {
		canisterID, err := pic.CreateCanister()
		if err != nil {
			t.Fatal(err)
		}

		wasmModule := compileMotoko(t, "assetstorage/actor.mo", "assetstorage/actor.wasm")
		if err := pic.InstallCode(*canisterID, wasmModule, nil, nil); err != nil {
			t.Fatal(err)
		}

		a, err := assetstorage.NewAgent(*canisterID, config)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := a.ApiVersion(); err != nil {
			t.Error(err)
		}

		if err := a.Authorize(principal.AnonymousID); err != nil {
			t.Fatal(err)
		}

		{
			a, err := agent.New(config)
			if err != nil {
				t.Fatal(err)
			}
			did, err := a.GetCanisterMetadata(*canisterID, "candid:service")
			if err != nil {
				t.Fatal(err)
			}
			if len(did) == 0 {
				t.Error("empty did")
			}
		}
	})

	t.Run("management canister", func(t *testing.T) {
		canisterID, err := pic.CreateCanister()
		if err != nil {
			t.Fatal(err)
		}

		wasmModule := compileMotoko(t, "ic/actor.mo", "ic/actor.wasm")
		if err := pic.InstallCode(*canisterID, wasmModule, nil, nil); err != nil {
			t.Fatal(err)
		}

		a, err := ic0.NewAgent(ic.MANAGEMENT_CANISTER_PRINCIPAL, config)
		if err != nil {
			t.Fatal(err)
		}

		if err := a.UpdateSettings(ic0.UpdateSettingsArgs{
			CanisterId: *canisterID,
			Settings: ic0.CanisterSettings{
				Controllers: &[]principal.Principal{
					principal.AnonymousID,
				},
			},
		}); err != nil {
			t.Error(err)
		}

		{ // Do the same manually.
			if err := a.Call(
				a.CanisterId,
				"update_settings",
				[]any{map[string]any{
					"canister_id": *canisterID,
					"settings": map[string]any{
						"controllers": &[]principal.Principal{
							principal.AnonymousID,
						},
					},
				}},
				[]any{},
			); err != nil {
				t.Error(err)
			}
		}

		t.Run("empty canister", func(t *testing.T) {
			a, err := agent.New(config)
			if err != nil {
				t.Fatal(err)
			}
			h, err := a.GetCanisterModuleHash(*canisterID)
			if err != nil {
				t.Fatal(err)
			}
			if len(h) != 32 {
				t.Error("hash length mismatch")
			}

			if err := pic.UninstallCode(*canisterID, nil); err != nil {
				t.Fatal(err)
			}

			h, err = a.GetCanisterModuleHash(*canisterID)
			if err != nil {
				t.Fatal(err)
			}
			if len(h) != 0 {
				t.Error("hash length mismatch")
			}
		})
	})
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

type localLogger struct{}

func (l localLogger) Printf(format string, v ...any) {
	fmt.Printf("[LOCAL]"+format+"\n", v...)
}
