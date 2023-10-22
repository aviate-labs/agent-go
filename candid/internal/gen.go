//go:build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func main() {
	for _, grammar := range []struct {
		path string
		name string
	}{
		{path: "internal/blob", name: "blob"},
		{path: "internal/candid"},
		{path: "internal/candidtest", name: "candidtest"},
		{path: "internal/candidvalue", name: "candidvalue"},
	} {
		rawGrammar, _ := ioutil.ReadFile(fmt.Sprintf("%s/grammar.pegn", grammar.path))
		if err := pegn.GenerateFromFiles(fmt.Sprintf("%s/", grammar.path), pegn.Config{
			ModulePath:     fmt.Sprintf("github.com/di-wu/candid-go/%s", grammar.path),
			ModuleName:     grammar.name,
			IgnoreReserved: true,
			TypeSuffix:     "T",
		}, rawGrammar); err != nil {
			log.Fatal(err)
		}
		log.Printf("Successfully generated the %s sub-module.\n", grammar.path)
	}
}
