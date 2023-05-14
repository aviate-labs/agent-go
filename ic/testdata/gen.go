package main

import (
	"embed"
	"fmt"
	"github.com/aviate-labs/agent-go/gen"
	"log"
	"os"
	"strings"
)

var (
	//go:embed did
	dids embed.FS
)

func main() {
	entries, _ := dids.ReadDir("did")
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".did")
		fmt.Printf("Generating %q...\n", name)
		did, _ := dids.ReadFile(fmt.Sprintf("did/%s", entry.Name()))

		dir := fmt.Sprintf("ic/%s", name)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.Mkdir(dir, os.ModePerm)
		}

		g, err := gen.NewGenerator(name, name, did)
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
