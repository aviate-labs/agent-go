package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/aviate-labs/agent-go/gen"
)

var (
	//go:embed did
	dids embed.FS

	ICVersion            = "release-2024-09-19_01-31-base"
	InterfaceSpecVersion = "0.23.0"
	SDKVersion           = "0.23.0"
)

func checkLatest() error {
	for _, f := range []struct {
		filepath string
		remote   string
	}{
		{
			filepath: "ic/testdata/did/assetstorage.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/sdk/%s/src/distributed/assetstorage.did", SDKVersion),
		},
		{
			filepath: "ic/testdata/did/cmc.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/nns/cmc/cmc.did", ICVersion),
		},
		{
			filepath: "ic/testdata/did/ic.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/interface-spec/%s/spec/_attachments/ic.did", InterfaceSpecVersion),
		},
		{
			filepath: "ic/testdata/did/registry.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/registry/canister/canister/registry.did", ICVersion),
		},
		{
			filepath: "ic/testdata/did/governance.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/nns/governance/canister/governance.did", ICVersion),
		},
		{
			filepath: "ic/testdata/did/icparchive.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/rosetta-api/icp_ledger/ledger_archive.did", ICVersion),
		},
		{
			filepath: "ic/testdata/did/icpledger.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/ic/%s/rs/rosetta-api/icp_ledger/ledger.did", ICVersion),
		},
		{
			filepath: "ic/testdata/did/icrc1.did",
			remote:   "https://raw.githubusercontent.com/dfinity/ICRC-1/master/standards/ICRC-1/ICRC-1.did",
		},
		{
			filepath: "ic/testdata/did/wallet.did",
			remote:   fmt.Sprintf("https://raw.githubusercontent.com/dfinity/sdk/%s/src/distributed/wallet.did", SDKVersion),
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

		if strings.HasSuffix(name, ".test") {
			name = strings.TrimSuffix(name, ".test")
			g, err := gen.NewGenerator(title(name), name, "ic_test", did)
			if err != nil {
				log.Panic(err)
			}
			raw, err := g.Generate()
			if err != nil {
				log.Panic(err)
			}
			_ = os.WriteFile(fmt.Sprintf("ic/%s_agent_test.go", name), raw, os.ModePerm)
		} else {
			dir := fmt.Sprintf("ic/%s", name)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				_ = os.Mkdir(dir, os.ModePerm)
			}

			{
				g, err := gen.NewGenerator("", name, name, did)
				if err != nil {
					log.Panic(err)
				}
				if name == "ic" {
					g.Indirect()
				}
				raw, err := g.Generate()
				if err != nil {
					log.Panic(err)
				}
				_ = os.WriteFile(fmt.Sprintf("%s/agent.go", dir), raw, os.ModePerm)
			}
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
