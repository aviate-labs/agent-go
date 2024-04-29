package pocketic

import (
	"encoding/json"
	"fmt"
)

type None struct{}

func (n None) MarshalJSON() ([]byte, error) {
	return json.Marshal("None")
}

func (n None) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	if s != "None" {
		return fmt.Errorf("expected None, got %s", s)
	}
	return nil
}
