package ledger

import (
	"fmt"
	"github.com/aviate-labs/agent-go/candid/idl"
	"net/url"
	"strconv"

	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/principal"
)

type Agent struct {
	a          agent.Agent
	canisterId principal.Principal
}

func New(canisterId principal.Principal, host *url.URL) Agent {
	return Agent{
		a: agent.New(agent.Config{
			ClientConfig: &agent.ClientConfig{Host: host},
		}),
		canisterId: canisterId,
	}
}

func (a Agent) AccountBalance(accountBalanceArgs AccountBalanceArgs) (*Tokens, error) {
	args, err := accountBalanceArgs.encode()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := a.a.Query(
		a.canisterId,
		"account_balance",
		args,
		[]any{&m},
	); err != nil {
		return nil, err
	}
	tokens, ok := recordTokens(m)
	if !ok {
		return nil, fmt.Errorf("invalid map: %s", m)
	}
	return tokens, nil
}

func (a Agent) Transfer(transferArgs TransferArgs) (*BlockIndex, error) {
	args, err := transferArgs.encode()
	if err != nil {
		return nil, err
	}
	var m *idl.Variant
	if err := a.a.Call(
		a.canisterId,
		"transfer",
		args,
		[]any{&m},
	); err != nil {
		return nil, err
	}
	i, _ := strconv.Atoi(m.Name)
	t := m.Type.(idl.VariantType).Fields[i]
	if ok := idl.HashString("Ok"); t.Name != ok {
		return nil, fmt.Errorf("ok (%s) not found, got: %s", ok, t.Name)
	}
	v := m.Value.(uint64)
	return &v, err
}
