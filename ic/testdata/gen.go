package main

import (
	"embed"
	"fmt"
	"github.com/aviate-labs/agent-go/gen"
	"log"
	"os"
	"strings"
	"unicode"
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
				raw, err := g.Generate()
				if err != nil {
					log.Panic(err)
				}
				_ = os.WriteFile(fmt.Sprintf("%s/agent.go", dir), raw, os.ModePerm)
			}
			{
				g, err := gen.NewGenerator("", name, name, did)
				g.ModulePath = "github.com/aviate-labs/agent-go/ic"
				if err != nil {
					log.Panic(err)
				}
				raw, err := g.GenerateMock()
				if err != nil {
					log.Panic(err)
				}
				_ = os.WriteFile(fmt.Sprintf("%s/agent_test.go", dir), raw, os.ModePerm)
			}
			{
				g, err := gen.NewGenerator("", name, name, did)
				g.ModulePath = "github.com/aviate-labs/agent-go/ic"
				if err != nil {
					log.Panic(err)
				}
				rawTypes, err := g.GenerateActorTypes()
				if err != nil {
					log.Panic(err)
				}
				_ = os.WriteFile(fmt.Sprintf("%s/types.mo", dir), rawTypes, os.ModePerm)
				rawActor, err := g.GenerateActor()
				if err != nil {
					log.Panic(err)
				}
				_ = os.WriteFile(fmt.Sprintf("%s/actor.mo", dir), rawActor, os.ModePerm)
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
