package main

import (
	"fmt"
	"os"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/cmd/goic/internal/cmd"
	"github.com/aviate-labs/agent-go/gen"
	"github.com/aviate-labs/agent-go/principal"
)

// rw-r--r-- : data files, not executable.
const outputPerm os.FileMode = 0o644

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
				Name:        "output",
				Description: "Write the DID to this file instead of stdout.",
				HasValue:    true,
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
				return os.WriteFile(path, rawDID, outputPerm)
			}
			fmt.Println(string(rawDID))
			return nil
		},
	),
	cmd.NewCommandFork(
		"generate",
		"Generate a new Agent from a DID file or a canister ID.",
		cmd.NewCommand(
			"did",
			"Generate a new Agent from a DID.",
			[]string{"path", "name"},
			[]cmd.CommandOption{
				{
					Name:        "output",
					Description: "Write the Agent to this file instead of stdout.",
					HasValue:    true,
				},
				{
					Name:        "packageName",
					Description: "Go package name for the generated code (default: name).",
					HasValue:    true,
				},
				{
					Name:        "agentName",
					Description: "Name of the generated Agent type (default: name).",
					HasValue:    true,
				},
				{
					Name:        "canisterID",
					Description: "Embed this canister ID in the generated Agent.",
					HasValue:    true,
				},
				{
					Name:        "indirect",
					Description: "Generate indirect (boxed) call wrappers.",
					HasValue:    false,
				},
			},
			func(args []string, options map[string]string) error {
				inputPath := args[0]
				o := parseGenOptions(args[1], options)

				var canisterID *principal.Principal
				if cID, ok := options["canisterID"]; ok {
					p, err := principal.Decode(cID)
					if err != nil {
						return err
					}
					canisterID = &p
				}

				g, err := gen.NewGeneratorFromFile(o.agentName, o.canisterName, o.packageName, inputPath)
				if err != nil {
					return err
				}
				return writeGenerated(g, canisterID, o.output, o.indirect)
			},
		),
		cmd.NewCommand(
			"remote",
			"Generate a new Agent from a canister ID.",
			[]string{"id", "canisterName"},
			[]cmd.CommandOption{
				{
					Name:        "output",
					Description: "Write the Agent to this file instead of stdout.",
					HasValue:    true,
				},
				{
					Name:        "packageName",
					Description: "Go package name for the generated code (default: canisterName).",
					HasValue:    true,
				},
				{
					Name:        "agentName",
					Description: "Name of the generated Agent type (default: canisterName).",
					HasValue:    true,
				},
				{
					Name:        "indirect",
					Description: "Generate indirect (boxed) call wrappers.",
					HasValue:    false,
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

				o := parseGenOptions(args[1], options)
				return writeDID(o.agentName, o.canisterName, &canisterID, o.packageName, o.output, []rune(string(rawDID)), o.indirect)
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

type genOptions struct {
	canisterName string
	packageName  string
	agentName    string
	output       string
	indirect     bool
}

// parseGenOptions reads the options shared by the generate subcommands.
// packageName and agentName default to canisterName when unset.
func parseGenOptions(canisterName string, options map[string]string) genOptions {
	o := genOptions{
		canisterName: canisterName,
		packageName:  canisterName,
		agentName:    canisterName,
		output:       options["output"],
	}
	if p, ok := options["packageName"]; ok {
		o.packageName = p
	}
	if a, ok := options["agentName"]; ok {
		o.agentName = a
	}
	_, o.indirect = options["indirect"]
	return o
}

func writeDID(agentName, canisterName string, canisterID *principal.Principal, packageName, outputPath string, rawDID []rune, indirect bool) error {
	g, err := gen.NewGenerator(agentName, canisterName, packageName, rawDID)
	if err != nil {
		return err
	}
	return writeGenerated(g, canisterID, outputPath, indirect)
}

func writeGenerated(g *gen.Generator, canisterID *principal.Principal, outputPath string, indirect bool) error {
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
		return os.WriteFile(outputPath, raw, outputPerm)
	}
	fmt.Println(string(raw))
	return nil
}
