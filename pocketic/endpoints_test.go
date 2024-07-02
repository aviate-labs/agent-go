package pocketic_test

import (
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/candid/idl"
	"github.com/aviate-labs/agent-go/pocketic"
	"github.com/aviate-labs/agent-go/principal"
	"net/url"
	"testing"
)

func TestEndpoints(t *testing.T) {
	pic, err := pocketic.New(
		pocketic.WithLogger(new(testLogger)),
		pocketic.WithNNSSubnet(),
		pocketic.WithApplicationSubnet(),
	)
	if err != nil {
		t.Skipf("skipping test: %v", err)
	}

	t.Run("status", func(t *testing.T) {
		if err := pic.Status(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("blobstore", func(t *testing.T) {
		id, err := pic.UploadBlob([]byte{0, 1, 2, 3}, false)
		if err != nil {
			t.Fatal(err)
		}
		bytes, err := pic.GetBlob(id)
		if err != nil {
			t.Fatal(err)
		}
		if len(bytes) != 4 {
			t.Fatalf("unexpected blob size: %d", len(bytes))
		}
	})
	t.Run("instances", func(t *testing.T) {
		var instances []string
		t.Run("get", func(t *testing.T) {
			instances, err = pic.GetInstances()
			if err != nil {
				t.Fatal(err)
			}
			if len(instances) == 0 {
				t.Fatal("no instances found")
			}
		})
		var instanceConfig *pocketic.InstanceConfig
		t.Run("post", func(t *testing.T) {
			instanceConfig, err = pic.CreateInstance(pocketic.DefaultSubnetConfig)
			if err != nil {
				t.Fatal(err)
			}
			if instanceConfig == nil {
				t.Fatal("instance config is nil")
			}
			newInstances, err := pic.GetInstances()
			if err != nil {
				t.Fatal(err)
			}
			if len(newInstances) != len(instances)+1 {
				t.Fatalf("unexpected instances count: %d", len(newInstances))
			}
		})
		t.Run("delete", func(t *testing.T) {
			if err := pic.DeleteInstance(instanceConfig.InstanceID); err != nil {
				t.Fatal(err)
			}
			newInstances, err := pic.GetInstances()
			if err != nil {
				t.Fatal(err)
			}
			if newInstances[len(newInstances)-1] != "Deleted" {
				t.Fatal("instance was not deleted")
			}
		})

		canisterID, err := pic.CreateCanister()
		if err != nil {
			t.Fatal(err)
		}
		wasmModule := compileMotoko(t, "testdata/main.mo", "testdata/main.wasm")
		if err := pic.InstallCode(*canisterID, wasmModule, nil, nil); err != nil {
			t.Fatal(err)
		}

		t.Run("query", func(t *testing.T) {
			if err := pic.QueryCall(*canisterID, principal.AnonymousID, "void", nil, nil); err == nil {
				t.Fatal()
			}
			var resp string
			if err := pic.QueryCall(*canisterID, principal.AnonymousID, "helloQuery", []any{"world"}, []any{&resp}); err != nil {
				t.Fatal(err)
			}
			if resp != "Hello, world!" {
				t.Fatalf("unexpected response: %s", resp)
			}
		})

		t.Run("update", func(t *testing.T) {
			if err := pic.UpdateCall(*canisterID, principal.AnonymousID, "void", nil, nil); err == nil {
				t.Fatal()
			}
			var resp string
			if err := pic.UpdateCall(*canisterID, principal.AnonymousID, "helloUpdate", []any{"world"}, []any{&resp}); err != nil {
				t.Fatal(err)
			}
			if resp != "Hello, world!" {
				t.Fatalf("unexpected response: %s", resp)
			}
		})

		t.Run("get_time", func(t *testing.T) {
			dt, err := pic.GetTime()
			if err != nil {
				t.Fatal(err)
			}
			if dt == nil {
				t.Fatal("time is nil")
			}
		})

		t.Run("get_cycles", func(t *testing.T) {
			cycles, err := pic.GetCycles(*canisterID)
			if err != nil {
				t.Fatal(err)
			}
			if cycles <= 0 {
				t.Fatalf("unexpected cycles: %d", cycles)
			}
		})

		t.Run("stable_memory", func(t *testing.T) {
			if err := pic.SetStableMemory(*canisterID, []byte{0, 1, 2, 3}, false); err != nil {
				t.Fatal(err)
			}
			if _, err := pic.GetStableMemory(*canisterID); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("get_subnet", func(t *testing.T) {
			subnetID, err := pic.GetSubnet(*canisterID)
			if err != nil {
				t.Fatal(err)
			}
			if subnetID == nil {
				t.Fatal("subnet ID is nil")
			}
		})

		t.Run("pub_key", func(t *testing.T) {
			if _, err := pic.RootKey(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("ingress_message", func(t *testing.T) {
			payload, err := idl.Marshal([]any{"world"})
			if err != nil {
				t.Fatal(err)
			}
			{
				msgID, err := pic.SubmitCall(*canisterID, principal.AnonymousID, "helloUpdate", payload)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := pic.AwaitCall(*msgID)
				if err != nil {
					t.Fatal(err)
				}
				var resp string
				if err := idl.Unmarshal(raw, []any{&resp}); err != nil {
					t.Fatal(err)
				}
				if resp != "Hello, world!" {
					t.Fatalf("unexpected response: %s", resp)
				}
			}
			{
				raw, err := pic.ExecuteCall(*canisterID, new(pocketic.EffectivePrincipalNone), principal.AnonymousID, "helloUpdate", payload)
				if err != nil {
					t.Fatal(err)
				}
				var resp string
				if err := idl.Unmarshal(raw, []any{&resp}); err != nil {
					t.Fatal(err)
				}
				if resp != "Hello, world!" {
					t.Fatalf("unexpected response: %s", resp)
				}
			}
		})

		t.Run("tick", func(t *testing.T) {
			if err := pic.Tick(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("api/v2/canister", func(t *testing.T) {
			host, err := url.Parse(pic.InstanceURL())
			if err != nil {
				t.Fatal(err)
			}
			a, err := NewAgent(*canisterID, agent.Config{
				ClientConfig: &agent.ClientConfig{Host: host},
				FetchRootKey: true,
				Logger:       new(testLogger),
			})
			if err != nil {
				t.Fatal(err)
			}

			if _, err := pic.MakeLive(nil); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := pic.MakeDeterministic(); err != nil {
					t.Fatal(err)
				}
			}()

			q, err := a.HelloQuery("world")
			if err != nil {
				t.Fatal(err)
			}
			if *q != "Hello, world!" {
				t.Fatalf("unexpected response: %s", *q)
			}

			u, err := a.HelloUpdate("world")
			if err != nil {
				t.Fatal(err)
			}
			if *u != "Hello, world!" {
				t.Fatalf("unexpected response: %s", *u)
			}
		})
	})
}
