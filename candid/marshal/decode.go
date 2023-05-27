package marshal

import (
	"fmt"
	"github.com/aviate-labs/agent-go/candid/idl"
)

func Unmarshal(data []byte, values []any) error {
	ts, vs, err := idl.Decode(data)
	if err != nil {
		return err
	}
	if len(ts) != len(vs) {
		return fmt.Errorf("unequal data types and value lengths: %d %d", len(ts), len(vs))
	}

	if len(vs) != len(values) {
		return fmt.Errorf("unequal value lengths: %d %d", len(vs), len(values))
	}

	for i, v := range values {
		if err := ts[i].UnmarshalGo(vs[i], v); err != nil {
			return err
		}
	}

	return nil
}
