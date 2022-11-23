package candid

import (
	"fmt"
)

func ExampleHashId() {
	fmt.Println(HashId("test"))
	fmt.Println(HashId("a_very_long_test"))
	// output:
	// 1291438162
	// 1731570314
}
