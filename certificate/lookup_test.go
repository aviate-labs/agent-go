package certificate_test

import (
	"fmt"

	"github.com/aviate-labs/agent-go/certificate"
)

func ExampleLookup() {
	fmt.Println(string(certificate.Lookup(certificate.LookupPath("a", "x"), tree)))
	fmt.Println(string(certificate.Lookup(certificate.LookupPath("a", "y"), tree)))
	fmt.Println(string(certificate.Lookup(certificate.LookupPath("b"), tree)))
	fmt.Println(string(certificate.Lookup(certificate.LookupPath("d"), tree)))
	// Output:
	// hello
	// world
	// good
	// morning
}
