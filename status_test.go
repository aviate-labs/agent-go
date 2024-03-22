package agent_test

import (
	"fmt"
	"net/url"

	"github.com/aviate-labs/agent-go"
)

var ic0URL, _ = url.Parse("https://ic0.app/")

func ExampleClient_Status() {
	c := agent.NewClient(agent.ClientConfig{Host: ic0URL})
	status, _ := c.Status()
	fmt.Println(status.Version)
	// Output:
	// 0.18.0
}
