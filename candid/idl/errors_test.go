package idl

import (
	"fmt"
)

func ExampleEncodeValueError() {
	fmt.Println(EncodeValueError{
		Expected: BoolOpCode,
		Value:    0,
	}.Error())
	// Output:
	// invalid type 0 (int), expected type bool
}
