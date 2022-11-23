package certificate_test

import (
	"fmt"

	cert "github.com/aviate-labs/agent-go/certificate"
)

func ExampleLookup() {
	fmt.Println(string(cert.Lookup(path("a", "x"), tree)))
	fmt.Println(string(cert.Lookup(path("a", "y"), tree)))
	fmt.Println(string(cert.Lookup(path("b"), tree)))
	fmt.Println(string(cert.Lookup(path("d"), tree)))
	// Output:
	// hello
	// world
	// good
	// morning
}

func path(p ...string) [][]byte {
	var path [][]byte
	for _, p := range p {
		path = append(path, []byte(p))
	}
	return path
}
