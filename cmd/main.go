package main

import (
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/cmd/internal/cmd"
	"github.com/aviate-labs/agent-go/gen"
	"github.com/aviate-labs/agent-go/principal"
	"os"
)

func fetchDID(canisterId principal.Principal) ([]byte, error) {
	a, err := agent.New(agent.Config{})
	if err != nil {
		return nil, err
	}
	var did string
	if err := a.Query(canisterId, "__get_candid_interface_tmp_hack", nil, []any{&did}); err != nil {
		return nil, err
	}
	return []byte(did), nil
}

func writeDID(canisterName, packageName, outputPath string, rawDID []byte) error {
	g, err := gen.NewGenerator(canisterName, packageName, rawDID)
	if err != nil {
		return err
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

var root = cmd.NewCommandFork(
	"agent-go",
	"agent-go is a CLI tool for creating a Go agent.",
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

				return writeDID(canisterName, packageName, path, rawDID)
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

				canisterName := args[1]
				packageName := canisterName
				if p, ok := options["packageName"]; ok {
					packageName = p
				}

				return writeDID(canisterName, packageName, path, rawDID)
			},
		),
	),
)

func main() {
	if err := root.Call(os.Args[1:]...); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
