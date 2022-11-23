package marshal_test

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/aviate-labs/agent-go/principal"
)

func hexToBytesReader(v string) *bytes.Reader {
	bs, _ := hex.DecodeString(v)
	return bytes.NewReader(bs)
}

func principalFromString(v string) principal.Principal {
	p, _ := principal.Decode(v)
	return p
}

func printDecode(val any, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(val)
	}
}

func printEncode(typ []byte, val []byte, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%x%x\n", typ, val)
	}
}
