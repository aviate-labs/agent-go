package mock_test

import (
	"bytes"
	"fmt"
	"github.com/aviate-labs/agent-go"
	"github.com/aviate-labs/agent-go/mock"
	"github.com/aviate-labs/agent-go/principal"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNewReplica(t *testing.T) {
	replica := mock.NewReplica()
	var canisterId principal.Principal
	replica.AddCanister(
		canisterId,
		func(request mock.Request) ([]any, error) {
			if !bytes.Equal(request.Sender.Raw, principal.AnonymousID.Raw) {
				t.Error("unexpected sender")
			}
			if request.MethodName != "test" {
				t.Error("unexpected method name")
			}
			if len(request.Arguments) != 0 {
				t.Error("unexpected arguments")
			}
			return []any{"hello"}, nil
		},
	)

	s := httptest.NewServer(replica)
	u, _ := url.Parse(s.URL)
	a, _ := agent.New(agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	})

	t.Run("call", func(t *testing.T) {
		var result string
		if err := a.Call(canisterId, "test", nil, []any{&result}); err != nil {
			t.Error(err)
		}
		if result != "hello" {
			t.Error("unexpected result")
		}
	})

	t.Run("query", func(t *testing.T) {
		var result string
		if err := a.Query(canisterId, "test", nil, []any{&result}); err != nil {
			t.Error(err)
		}
		if result != "hello" {
			t.Error("unexpected result")
		}
	})
}

func TestNewReplica_error(t *testing.T) {
	replica := mock.NewReplica()
	var canisterId principal.Principal
	replica.AddCanister(
		canisterId,
		func(request mock.Request) ([]any, error) {
			return nil, fmt.Errorf("oops")
		},
	)

	s := httptest.NewServer(replica)
	u, _ := url.Parse(s.URL)
	a, _ := agent.New(agent.Config{
		ClientConfig: &agent.ClientConfig{Host: u},
		FetchRootKey: true,
	})

	t.Run("call", func(t *testing.T) {
		var result string
		err := a.Call(canisterId, "test", nil, []any{&result})
		if err == nil || err.Error() != "(500) 500 Internal Server Error: oops" {
			t.Error("unexpected error")
		}
	})
}
