package idl_test

import (
	"testing"

	"github.com/aviate-labs/agent-go/candid/idl"
)

// hash(id) = ( Sum_(i=0..k) utf8(id)[i] * 223^(k-i) ) mod 2^32: the sum is over
// utf8 bytes, so a non-ascii name does not hash as runes.
func TestHashUTF8Bytes(t *testing.T) {
	// dfinity/candid construct.test.did "variant: unicode field" carries tag
	// 11272781 for "☃".
	if got := idl.HashString("☃"); got != "11272781" {
		t.Fatalf("hash(☃) = %s, want 11272781", got)
	}
	if got := idl.HashString("foo"); got != "5097222" {
		t.Fatalf("hash(foo) = %s, want 5097222", got)
	}
}
