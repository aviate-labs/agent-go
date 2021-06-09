package crc32_test

import (
	"encoding/hex"
	"fmt"

	"github.com/allusion-be/agent-go/internal/crc32"
)

// Source: RFC 3385
// https://www.rfc-editor.org/rfc/rfc3385

func Example_vector() {
	b, _ := hex.DecodeString("ABCD01")
	h := crc32.New(b)
	fmt.Printf("%x\n", h.Value())
	// Output:
	// 233ff206
}
