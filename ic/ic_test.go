package ic_test

import (
	"encoding/json"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/ic"
	"github.com/aviate-labs/agent-go/ic/assetstorage"
	ic0 "github.com/aviate-labs/agent-go/ic/ic"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestModules(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	rawNetworksConfig, err := os.ReadFile(fmt.Sprintf("%s/.config/dfx/networks.json", homeDir))
	if err != nil {
		t.Skip(err)
	}
	var networksConfig networkConfig
	if err := json.Unmarshal(rawNetworksConfig, &networksConfig); err != nil {
		t.Fatal(err)
	}
	host, err := url.Parse(fmt.Sprintf("http://%s", networksConfig.Local.Bind))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Using DFX host:", host)

	dfxPath, err := exec.LookPath("dfx")
	if err != nil {
		t.Skip(err)
	}
	start := exec.Command(dfxPath, "start", "--background", "--clean")
	if err := start.Start(); err != nil {
		t.Fatal(err)
	}
	if err := start.Wait(); err != nil {
		t.Error(err)
	}
	t.Log("Started DFX")
	defer func() {
		out, _ := exec.Command(dfxPath, "stop").CombinedOutput()
		t.Log(sanitizeOutput(out))
	}()

	deploy := exec.Command(dfxPath, "deploy", "--no-wallet")
	if out, err := deploy.CombinedOutput(); err != nil {
		t.Fatal(sanitizeOutput(out))
	}

	raw, err := os.ReadFile(".dfx/local/canister_ids.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}

	config := agent.Config{
		ClientConfig: &agent.ClientConfig{Host: host},
		FetchRootKey: true,
		Logger:       new(localLogger),
	}

	t.Run("assetstorage", func(t *testing.T) {
		cId, _ := principal.Decode(m["assetstorage"]["local"])
		a, err := assetstorage.NewAgent(cId, config)
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
			did, err := a.GetCanisterMetadata(cId, "candid:service")
			if err != nil {
				t.Fatal(err)
			}
			if len(did) == 0 {
				t.Error("empty did")
			}
		}
	})

	t.Run("management canister", func(t *testing.T) {
		controller := principal.AnonymousID
		addController := exec.Command(dfxPath, "canister", "update-settings", "--add-controller", controller.String(), "ic0")
		if out, err := addController.CombinedOutput(); err != nil {
			t.Fatal(sanitizeOutput(out))
		}

		getContollers := exec.Command(dfxPath, "canister", "info", "ic0")
		out, err := getContollers.CombinedOutput()
		if err != nil {
			t.Fatal(sanitizeOutput(out))
		}
		if !strings.Contains(string(out), controller.String()) {
			t.Error("controller not added")
		}

		cId, _ := principal.Decode(m["ic0"]["local"])
		a, err := ic0.NewAgent(ic.MANAGEMENT_CANISTER_PRINCIPAL, config)
		if err != nil {
			t.Fatal(err)
		}

		if err := a.UpdateSettings(ic0.UpdateSettingsArgs{
			CanisterId: cId,
			Settings: ic0.CanisterSettings{
				Controllers: &[]principal.Principal{
					principal.AnonymousID,
				},
			},
		}); err != nil {
			t.Error(err)
		}

		t.Run("empty canister", func(t *testing.T) {
			a, err := agent.New(config)
			if err != nil {
				t.Fatal(err)
			}
			h, err := a.GetCanisterModuleHash(cId)
			if err != nil {
				t.Fatal(err)
			}
			if len(h) != 32 {
				t.Error("hash length mismatch")
			}

			uninstall := exec.Command(dfxPath, "canister", "uninstall-code", "ic0", "--identity", "anonymous")
			if out, err := uninstall.CombinedOutput(); err != nil {
				t.Fatal(sanitizeOutput(out))
			}

			h, err = a.GetCanisterModuleHash(cId)
			if err != nil {
				t.Fatal(err)
			}
			if len(h) != 0 {
				t.Error("hash length mismatch")
			}
		})
	})
}

func sanitizeOutput(out []byte) string {
	const artifact = "\u001B(B" // Not sure where this comes from...
	var s string
	for _, p := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		s += strings.TrimSpace(strings.ReplaceAll(p, artifact, "")) + "\n"
	}
	return s
}

type localLogger struct{}

func (l localLogger) Printf(format string, v ...any) {
	fmt.Printf("[LOCAL]"+format+"\n", v...)
}

type networkConfig struct {
	Local struct {
		Bind string `json:"bind"`
	} `json:"local"`
}
