package main

import (
	"fmt"
	"os"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/cmd/goic/internal/cmd"
	"github.com/aviate-labs/agent-go/gen"
	"github.com/aviate-labs/agent-go/principal"
)

var root = cmd.NewCommandFork(
	"goic",
	"`goic` is a CLI tool for creating a Go agent.",
	cmd.NewCommand(
		"version",
		"Print the version of `goic`.",
		[]string{},
		[]cmd.CommandOption{},
		func(args []string, options map[string]string) error {
			fmt.Println("0.0.1")
			return nil
		},
	),
	cmd.NewCommand(
		"fetch",
		"Fetch a DID from a canister ID.",
		[]string{"id"},
		[]cmd.CommandOption{
			{
				Name:     "output",
				HasValue: true,
			},
		},
		func(args []string, options map[string]string) error {
			id := args[0]
			canisterId, err := principal.Decode(id)
			if err != nil {
				return err
			}
			rawDID, err := fetchDID(canisterId)
			if err != nil {
				return err
			}

			var path string
			if p, ok := options["output"]; ok {
				path = p
			}
			if path != "" {
				return os.WriteFile(path, rawDID, os.ModePerm)
			}
			fmt.Println(string(rawDID))
			return nil
		},
	),
	cmd.NewCommandFork(
		"generate",
		"Generate a new Agent from...",
		cmd.NewCommand(
			"did",
			"Generate a new Agent from a DID.",
			[]string{"path", "name"},
			[]cmd.CommandOption{
				{
					Name:     "output",
					HasValue: true,
				},
				{
					Name:     "packageName",
					HasValue: true,
				},
				{
					Name:     "agentName",
					HasValue: true,
				},
				{
					Name:     "canisterID",
					HasValue: true,
				},
				{
					Name:     "indirect",
					HasValue: false,
				},
			},
			func(args []string, options map[string]string) error {
				inputPath := args[0]
				rawDID, err := os.ReadFile(inputPath)
				if err != nil {
					return err
				}

				var path string
				if p, ok := options["output"]; ok {
					path = p
				}

				canisterName := args[1]
				packageName := canisterName
				if p, ok := options["packageName"]; ok {
					packageName = p
				}

				var agentName string
				if a, ok := options["agentName"]; ok {
					agentName = a
				}

				var canisterID *principal.Principal
				if cID, ok := options["canisterID"]; ok {
					p, err := principal.Decode(cID)
					if err != nil {
						return err
					}
					canisterID = &p
				}

				_, indirect := options["indirect"]
				return writeDID(agentName, canisterName, canisterID, packageName, path, []rune(string(rawDID)), indirect)
			},
		),
		cmd.NewCommand(
			"remote",
			"Generate a new Agent from a canister ID.",
			[]string{"id", "canisterName"},
			[]cmd.CommandOption{
				{
					Name:     "output",
					HasValue: true,
				},
				{
					Name:     "packageName",
					HasValue: true,
				},
				{
					Name:     "agentName",
					HasValue: true,
				},
				{
					Name:     "indirect",
					HasValue: false,
				},
			},
			func(args []string, options map[string]string) error {
				id := args[0]
				canisterID, err := principal.Decode(id)
				if err != nil {
					return err
				}
				rawDID, err := fetchDID(canisterID)
				if err != nil {
					return err
				}

				var path string
				if p, ok := options["output"]; ok {
					path = p
				}

				canisterName := args[1]
				packageName := canisterName
				if p, ok := options["packageName"]; ok {
					packageName = p
				}

				var agentName string
				if a, ok := options["agentName"]; ok {
					agentName = a
				}

				_, indirect := options["indirect"]
				return writeDID(agentName, canisterName, &canisterID, packageName, path, []rune(string(rawDID)), indirect)
			},
		),
	),
)

func fetchDID(canisterId principal.Principal) ([]byte, error) {
	a, err := agent.New(agent.Config{})
	if err != nil {
		return nil, err
	}
	var did string
	// This endpoint has been deprecated and removed starting with moc v0.11.0.
	if err := a.Query(canisterId, "__get_candid_interface_tmp_hack", nil, []any{&did}); err != nil {
		// It is recommended for the canister to have a custom section called "icp:public candid:service", which
		// contains the UTF-8 encoding of the Candid interface for the canister.
		return a.GetCanisterMetadata(canisterId, "candid:service")
	}
	return []byte(did), nil
}

func main() {
	if err := root.Call(os.Args[1:]...); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}

func writeDID(agentName, canisterName string, canisterID *principal.Principal, packageName, outputPath string, rawDID []rune, indirect bool) error {
	g, err := gen.NewGenerator(agentName, canisterName, packageName, rawDID)
	if err != nil {
		return err
	}
	if indirect {
		g.Indirect()
	}
	if canisterID != nil {
		g.WithCanisterID(canisterID)
	}
	raw, err := g.Generate()
	if err != nil {
		return err
	}

	if outputPath != "" {
		return os.WriteFile(outputPath, raw, os.ModePerm)
	}
	fmt.Println(string(raw))
	return nil
}
