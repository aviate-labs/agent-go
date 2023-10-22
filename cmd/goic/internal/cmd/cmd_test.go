package cmd

import (
	"fmt"
)

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
	fmt.Println(root.Call("a"))
	fmt.Println(root.Call("b"))
	// Output:
	// <nil>
	// command "" not found
	// command "b" not found
}
