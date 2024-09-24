package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/aviate-labs/agent-go/gen"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

var (
	//go:embed did
	dids embed.FS

	ICVersion  = "release-2024-09-19_01-31-base"
	SDKVersion = "0.23.0"
)

func checkLatest() error {
	for _, f := range []struct {
		filepath string
		remote   string
	}{
		{
			filepath: "ic/sns/testdata/did/sns.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/nns/sns-wasm/canister/sns-wasm.did", ICVersion),
		},
		{
			filepath: "ic/sns/testdata/did/governance.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/sdk/%s/src/distributed/assetstorage.did", SDKVersion),
		},
		{
			filepath: "ic/sns/testdata/did/root.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/sns/root/canister/root.did", ICVersion),
		},
		{
			filepath: "ic/sns/testdata/did/swap.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/sns/swap/canister/swap.did", ICVersion),
		},
		{
			filepath: "ic/sns/testdata/did/ledger.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/rosetta-api/icrc1/ledger/ledger.did", ICVersion),
		},
		{
			filepath: "ic/sns/testdata/did/index.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/rosetta-api/icrc1/index/index.did", ICVersion),
		},
	} {
		raw, err := http.Get(f.remote)
		if err != nil {
			return err
		}
		remoteDID, err := io.ReadAll(raw.Body)
		if err != nil {
			return err
		}
		localDID, err := os.ReadFile(f.filepath)
		if err != nil {
			return err
		}
		if bytes.Compare(remoteDID, localDID) != 0 {
			if err := os.WriteFile(f.filepath, remoteDID, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	if err := checkLatest(); err != nil {
		log.Panic(err)
	}

	entries, _ := dids.ReadDir("did")
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".did")
		fmt.Printf("Generating %q...\n", name)
		did, _ := dids.ReadFile(fmt.Sprintf("did/%s", entry.Name()))
		dir := fmt.Sprintf("ic/sns/%s", name)
		if name == "sns" {
			dir = "ic/sns"
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.Mkdir(dir, os.ModePerm)
		}

		{
			g, err := gen.NewGenerator("", name, name, did)
			if err != nil {
				log.Panic(err)
			}
			raw, err := g.Generate()
			if err != nil {
				log.Panic(err)
			}
			_ = os.WriteFile(fmt.Sprintf("%s/agent.go", dir), raw, os.ModePerm)
		}
	}
}

func title(s string) string {
	var title []rune
	for i, c := range s {
		if i == 0 {
			title = append(title, unicode.ToUpper(c))
		} else {
			title = append(title, c)
		}
	}
	return string(title)
}
