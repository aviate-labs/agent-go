package did

import (
	"fmt"

	"github.com/aviate-labs/agent-go/candid/internal/candid"
	"github.com/di-wu/parser/ast"
)

// Method is a public method of a service.
type Method struct {
	// Name describes the method.
	Name string

	// Func is a function type describing its signature.
	Func *Func
	// Id is a reference to a type definition naming a function reference type.
	// It is NOT possible to have both a function type and a reference.
	Id *string
}

func (m Method) String() string {
	s := fmt.Sprintf("%s : ", m.Name)
	if id := m.Id; id != nil {
		return s + *id
	}
	return s + m.Func.String()
}

// Service can be used to declare the complete interface of a service. A service is a standalone actor on the platform
// that can communicate with other services via sending and receiving messages. Messages are sent to a service by
// invoking one of its methods, i.e., functions that the service provides.
//
// Example:
//
//	service : {
//		addUser : (name : text, age : nat8) -> (id : nat64);
//		userName : (id : nat64) -> (text) query;
//		userAge : (id : nat64) -> (nat8) query;
//		deleteUser : (id : nat64) -> () oneway;
//	}
type Service struct {
	// Id represents the optional name given to the service. This only serves as documentation.
	Id *string

	// Methods is the list of methods that the service provides.
	Methods []Method
	// MethodId is the reference to the name of a type definition for an actor reference type.
	// It is NOT possible to have both a list of methods and a reference.
	MethodId *string
}

func convertService(n *ast.Node) Service {
	var actor Service
	for _, n := range n.Children() {
		switch n.Type {
		case candid.IdT:
			id := n.Value
			if actor.Id == nil {
				actor.Id = &id
				continue
			}
			actor.MethodId = &id
		case candid.TupTypeT:
			// TODO: what does this even do?
		case candid.ActorTypeT:
			for _, n := range n.Children() {
				name := n.FirstChild.Value
				switch n := n.LastChild; n.Type {
				case candid.FuncTypeT:
					f := convertFunc(n)
					actor.Methods = append(
						actor.Methods,
						Method{
							Name: name,
							Func: &f,
						},
					)
				case candid.IdT, candid.TextT:
					id := n.Value
					actor.Methods = append(
						actor.Methods,
						Method{
							Name: name,
							Id:   &id,
						},
					)
				default:
					panic(n)
				}
			}
		default:
			panic(n)
		}
	}
	return actor
}

func (a Service) String() string {
	s := "service "
	if id := a.Id; id != nil {
		s += fmt.Sprintf("%s ", *id)
	}
	s += ": "
	if id := a.MethodId; id != nil {
		return s + *id
	}
	s += "{\n"
	for _, m := range a.Methods {
		s += fmt.Sprintf("  %s;\n", m.String())
	}
	return s + "}"
}
