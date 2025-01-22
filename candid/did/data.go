package did

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/0x51-dev/upeg/parser"
	"github.com/aviate-labs/agent-go/candid/internal/candid"
)

func convertNat(n *parser.Node) *big.Int {
	switch n := strings.ReplaceAll(n.Value(), "_", ""); {
	case strings.HasPrefix(n, "0x"):
		n = strings.TrimPrefix(n, "0x")
		i, _ := strconv.ParseInt(n, 16, 64)
		return big.NewInt(i)
	default:
		i, _ := strconv.ParseInt(n, 10, 64)
		return big.NewInt(i)
	}
}

// Blob can be used for binary data, that is, sequences of bytes.
type Blob struct{}

func (b Blob) String() string {
	return "blob"
}
func (b Blob) data() {}

// Data is the content of message arguments and results.
type Data interface {
	data()
	fmt.Stringer
}

func convertData(n *parser.Node) Data {
	switch n.Name {
	case candid.Blob.Name:
		return Blob{}
	case candid.Opt.Name:
		return Optional{
			Data: convertData(n.Children()[0]),
		}
	case candid.Vec.Name:
		return Vector{
			Data: convertData(n.Children()[0]),
		}
	case candid.Record.Name:
		var record Record
		for _, n := range n.Children() {
			if n.Name == candid.CommentText.Name {
				continue
			}
			record = append(
				record,
				convertField(n),
			)
		}
		return record
	case candid.Variant.Name:
		var variant Variant
		for _, n := range n.Children() {
			if n.Name == candid.CommentText.Name {
				continue
			}
			variant = append(
				variant,
				convertField(n),
			)
		}
		return variant
	case candid.Func.Name:
		return convertFunc(n.Children()[0])
	case candid.Service.Name:
		return convertService(n.Children()[0])
	case candid.Principal.Name:
		return Principal{}
	case candid.PrimType.Name:
		return Primitive(n.Value())
	case candid.Id.Name:
		return DataId(n.Value())
	default:
		panic(n.Name)
	}
}

// DataId is an id reference to a data type.
type DataId string

func (i DataId) String() string {
	return string(i)
}
func (i DataId) data() {}

// Field
// The order in which fields are specified is immaterial.
type Field struct {
	// Nat is the field id.
	// e.g. 0 : nat
	Nat *big.Int
	// Name is the name of the field.
	// e.g. name : text
	Name *string

	// Data is a single value of specified data type that is carried.
	Data *Data

	// Only in variants.
	NatData  *big.Int
	NameData *string
}

func convertField(n *parser.Node) Field {
	var field Field
	if len(n.Children()) != 1 {
		switch n := n.Children()[0]; n.Name {
		case candid.Nat.Name:
			field.Nat = convertNat(n)
		case candid.Text.Name, candid.Id.Name:
			v := n.Value()
			field.Name = &v
		default:
			panic(n)
		}
	}
	switch n := n.Children()[0]; n.Name {
	case candid.Nat.Name:
		field.NatData = convertNat(n)
	case candid.Id.Name:
		v := n.Value()
		field.NameData = &v
	default:
		data := convertData(n)
		field.Data = &data
	}
	return field
}

func (f Field) String() string {
	var s string
	if n := f.Nat; n != nil {
		s += fmt.Sprintf("%s : ", n.String())
	} else if f.Name != nil {
		s += fmt.Sprintf("%s : ", *f.Name)
	}
	if f.Data != nil {
		d := *f.Data
		s += d.String()
	} else if n := f.NatData; n != nil {
		s += n.String()
	} else {
		s += *f.NameData
	}
	return s
}

func (f Func) data() {}

// Optional is used to express that some value is optional, meaning that data might
// be present as some value of type t, or might be absent as the value null.
type Optional struct {
	Data Data
}

func (o Optional) String() string {
	return fmt.Sprintf("opt %s", o.Data.String())
}

func (o Optional) data() {}

// Primitive describes the possible forms of primitive data.
type Primitive string

func (p Primitive) String() string {
	return string(p)
}

func (p Primitive) data() {}

// Principal is the common scheme to identify canisters, users, and other entities.
type Principal struct{}

func (p Principal) String() string {
	return "principal"
}

func (p Principal) data() {}

// Record a collection of labeled values.
type Record []Field

func (r Record) String() string {
	s := "record {\n"
	for _, f := range r {
		s += fmt.Sprintf("  %s;\n", f.String())
	}
	return s + "}"
}

func (r Record) data() {}

func (a Service) data() {}

// Variant represents a value that is from exactly one of the given cases, or tags.
type Variant []Field

func (v Variant) String() string {
	s := "variant {\n"
	for _, f := range v {
		s += fmt.Sprintf("  %s;\n", f.String())
	}
	return s + "}"
}

func (v Variant) data() {}

// Vector represents vectors (sequences, lists, arrays).
// e.g. 'vec bool', 'vec nat8', 'vec vec text', etc
type Vector struct {
	Data Data
}

func (v Vector) String() string {
	return fmt.Sprintf("vec %s", v.Data.String())
}

func (v Vector) data() {}
