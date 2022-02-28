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
		t.Skip("DFX not installed.")
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
		ps, err := agent.GetCanisterControllers(canister)
		if err != nil {
			t.Fatal(err)
		}
		if len(ps) != 1 {
			t.Fatal()
		}
		if p := ps[0].Encode(); p != "uea77-ug7xt-mi62f-fobao-tkelf-qjqxl-v62ed-rgqfd-oylqe-4l5xa-sae" {
			t.Error(p)
		}
	}
	{
		mh, err := agent.GetCanisterModuleHash(canister)
		if err != nil {
			t.Fatal(err)
		}
		if h := fmt.Sprintf("%x", mh); h != "b3d95eb1b6ddcc240afe7c79a2e05fb8e832f72019273fbc447a38f4ea651d56" {
			t.Error(h)
		}
	}
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
	{
		args, _ := candid.EncodeValue("( 1 : nat )")
		resp, err := agent.Call(canister, "add", args)
		if err != nil {
			t.Fatal(err)
		}
		if resp != "(1 : nat)" {
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
	time.Sleep(3 * time.Second)

	deploy := exec.Command(path, "deploy")
	deploy.Dir = "./testdata"
	if deploy.Run(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	controllers := exec.Command(path, "canister", "update-settings", "main",
		"--controller", "uea77-ug7xt-mi62f-fobao-tkelf-qjqxl-v62ed-rgqfd-oylqe-4l5xa-sae",
	)
	controllers.Dir = "./testdata"
	if err := controllers.Run(); err != nil {
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
