package agent_test

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"testing"
	"time"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/identity"
	"github.com/aviate-labs/candid-go"
	"github.com/aviate-labs/principal-go"
)

func TestLocalReplica(t *testing.T) {
	if _, err := exec.LookPath("dfx"); err != nil {
		t.Skip()
		return
	}

	cmd := startDFX(t)
	defer stopDFX(cmd, t)

	ic0, _ := url.Parse("http://localhost:8000/")
	canister, _ := principal.Decode("rrkah-fqaaa-aaaaa-aaaaq-cai")

	data, _ := ioutil.ReadFile("./testdata/test.pem")
	var id identity.Identity
	id, _ = identity.NewEd25519IdentityFromPEM(data)
	agent := agent.New(agent.AgentConfig{
		Identity: &id,
		ClientConfig: &agent.ClientConfig{
			Host: ic0,
		},
	})
	{
		args, _ := candid.EncodeValue("()")
		resp, err := agent.Query(canister, "get", args)
		if err != nil {
			t.Fatal(err)
		}
		if resp != "(0 : nat)" {
			t.Error(resp)
		}
	}
}

func startDFX(t *testing.T) *exec.Cmd {
	path, err := exec.LookPath("dfx")
	if err != nil {
		t.Fatal(err)
	}
	dfx := exec.Command(path, "start", "--background", "--clean")
	dfx.Dir = "./testdata"
	if err := dfx.Start(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Starting DFX...")
	time.Sleep(5 * time.Second)

	deploy := exec.Command(path, "deploy")
	deploy.Dir = "./testdata"
	if deploy.Run(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	controllers := exec.Command(path, "canister", "update-settings", "main",
		"--controller=\"$(dfx identity get-principal)\"",
		"--controller=\"uea77-ug7xt-mi62f-fobao-tkelf-qjqxl-v62ed-rgqfd-oylqe-4l5xa-sae\"",
	)
	controllers.Dir = "./testdata"
	if deploy.Run(); err != nil {
		t.Fatal(err)
	}

	return dfx
}

func stopDFX(dfx *exec.Cmd, t *testing.T) {
	path, err := exec.LookPath("dfx")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(path, "stop")
	cmd.Dir = "./testdata"
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if err := dfx.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("Stopped DFX.")
}
