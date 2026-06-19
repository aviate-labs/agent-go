package cmd

import (
	"fmt"
)

func ExampleCommandFork_Usage() {
	a := NewCommand("did", "Generate from a DID.", nil, nil,
		func(args []string, options map[string]string) error { return nil })
	b := NewCommand("remote", "Generate from a canister ID.", nil, nil,
		func(args []string, options map[string]string) error { return nil })
	root := NewCommandFork("generate", "Generate a new Agent.", a, b)
	fmt.Println(root.Usage())
	// Output:
	// generate <command>
	//   Generate a new Agent.
	//
	// Commands:
	//   did     Generate from a DID.
	//   remote  Generate from a canister ID.
}

func ExampleCommandFork_help() {
	a := NewCommand("did", "Generate from a DID.", nil, nil,
		func(args []string, options map[string]string) error { return nil })
	root := NewCommandFork("generate", "Generate a new Agent.", a)
	fmt.Println(root.Call())
	fmt.Println(root.Call("help"))
	// Output:
	// generate <command>
	//   Generate a new Agent.
	//
	// Commands:
	//   did  Generate from a DID.
	// <nil>
	// generate <command>
	//   Generate a new Agent.
	//
	// Commands:
	//   did  Generate from a DID.
	// <nil>
}

func ExampleCommand_Usage() {
	c := NewCommand(
		"did", "Generate a new Agent from a DID.",
		[]string{"path", "name"},
		[]CommandOption{
			{Name: "output", Description: "Output file.", HasValue: true},
			{Name: "indirect", Description: "Use indirect calls."},
		},
		func(args []string, options map[string]string) error { return nil },
	)
	fmt.Println(c.Usage())
	// Output:
	// did <path> <name> [options]
	//   Generate a new Agent from a DID.
	//
	// Options:
	//   --output=<value>  Output file.
	//   --indirect        Use indirect calls.
}

func ExampleInternalCommand() {
	b := NewCommand(
		"b", "b",
		nil, nil,
		func(args []string, options map[string]string) error {
			return nil
		},
	)
	a := NewCommandFork("a", "a", b)
	root := NewCommandFork("root", "root", a)
	fmt.Println(root.Call("a", "b"))
	fmt.Println(root.Call("b"))
	// Output:
	// <nil>
	// command "b" not found
}
