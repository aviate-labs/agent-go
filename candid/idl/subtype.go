package idl

import (
	"fmt"
	"math/big"
	"slices"
)

// Subtype reports whether a value of type t can be read at type u, i.e. t <: u
// per the candid spec's "Upgrading and Subtyping".
func Subtype(t, u Type) error {
	return subtype(t, u, nil)
}

type typePair struct{ t, u Type }

// subtype tracks the pairs it is already checking further up the stack. A
// recursive type would otherwise loop forever: seeing the same pair twice means
// the answer depends on itself, so take it as holding.
func subtype(t, u Type, gamma []typePair) error {
	t, u = unwrapRec(t), unwrapRec(u)
	if t == u {
		return nil
	}
	for _, g := range gamma {
		if g.t == t && g.u == u {
			return nil
		}
	}
	gamma = append(gamma, typePair{t, u})

	// reserved is top, empty is bottom.
	if _, ok := u.(*ReservedType); ok {
		return nil
	}
	if _, ok := t.(*EmptyType); ok {
		return nil
	}

	switch u := u.(type) {
	case *OptionalType:
		return subtypeOpt(t, u, gamma)
	case *VectorType:
		v, ok := t.(*VectorType)
		if !ok {
			return notSubtype(t, u)
		}
		return subtype(v.Type, u.Type, gamma)
	case *RecordType:
		return subtypeRecord(t, u, gamma)
	case *VariantType:
		return subtypeVariant(t, u, gamma)
	case *FunctionType:
		return subtypeFunc(t, u, gamma)
	case *ServiceType:
		return subtypeService(t, u, gamma)
	case *PrincipalType:
		if _, ok := t.(*ServiceType); ok {
			return nil
		}
	}

	if isNat(t) && isInt(u) {
		return nil
	}
	if t.String() == u.String() {
		return nil
	}
	return notSubtype(t, u)
}

func subtypeOpt(t Type, u *OptionalType, gamma []typePair) error {
	if _, ok := t.(*NullType); ok {
		return nil
	}
	// "any type is a subtype of an option": an unreadable constituent is seen as
	// null rather than an error, so opt-vs-opt always holds.
	if _, ok := t.(*OptionalType); ok {
		return nil
	}
	if err := subtype(t, u.Type, gamma); err == nil && !acceptsNull(u.Type) {
		return nil
	}
	// Promoting t into the option means reading it at the constituent type. If
	// that is options all the way down, as in `type Opt = opt Opt`, there is no
	// value to read it as.
	if optDepth(u, 0) < 0 {
		return notSubtype(t, u)
	}
	return nil
}

// optDepth counts nested options, returning -1 when there is no end to them.
// A depth cap would reject a merely deep chain, so the chain is walked until it
// either ends or revisits a type it has already been through.
func optDepth(t Type, n int) int {
	return optDepthSeen(t, n, nil)
}

func optDepthSeen(t Type, n int, seen []Type) int {
	o, ok := unwrapRec(t).(*OptionalType)
	if !ok {
		return n
	}
	if slices.Contains(seen, Type(o)) {
		return -1
	}
	return optDepthSeen(o.Type, n+1, append(seen, o))
}

func subtypeRecord(t Type, u *RecordType, gamma []typePair) error {
	r, ok := t.(*RecordType)
	if !ok {
		return notSubtype(t, u)
	}
	have := make(map[string]Type, len(r.Fields))
	for _, f := range r.Fields {
		have[LabelHash(f.Name)] = f.Type
	}
	for _, f := range u.Fields {
		ft, ok := have[LabelHash(f.Name)]
		if !ok {
			if !acceptsNull(f.Type) {
				return fmt.Errorf("record field %s: missing and not optional", f.Name)
			}
			continue
		}
		if err := subtype(ft, f.Type, gamma); err != nil {
			return fmt.Errorf("record field %s: %w", f.Name, err)
		}
	}
	return nil
}

// subtypeVariant only checks tags common to both: a decoder reads one tag, so
// an unknown wire tag is unreachable rather than an error.
func subtypeVariant(t Type, u *VariantType, gamma []typePair) error {
	v, ok := t.(*VariantType)
	if !ok {
		return notSubtype(t, u)
	}
	want := make(map[string]Type, len(u.Fields))
	for _, f := range u.Fields {
		want[LabelHash(f.Name)] = f.Type
	}
	for _, f := range v.Fields {
		ft, ok := want[LabelHash(f.Name)]
		if !ok {
			continue
		}
		if err := subtype(f.Type, ft, gamma); err != nil {
			return fmt.Errorf("variant tag %s: %w", f.Name, err)
		}
	}
	return nil
}

// subtypeFunc: arguments are contravariant, results covariant.
func subtypeFunc(t Type, u *FunctionType, gamma []typePair) error {
	f, ok := t.(*FunctionType)
	if !ok {
		return notSubtype(t, u)
	}
	if !sameAnnotations(f, u) {
		return fmt.Errorf("function annotations differ")
	}
	if err := subtypeParams(u.ArgumentParameters, f.ArgumentParameters, gamma); err != nil {
		return fmt.Errorf("function argument: %w", err)
	}
	if err := subtypeParams(f.ReturnParameters, u.ReturnParameters, gamma); err != nil {
		return fmt.Errorf("function result: %w", err)
	}
	return nil
}

// subtypeParams treats a parameter list as a tuple record.
func subtypeParams(ts, us []FunctionParameter, gamma []typePair) error {
	for i, u := range us {
		if i >= len(ts) {
			if !acceptsNull(u.Type) {
				return fmt.Errorf("parameter %d: missing and not optional", i)
			}
			continue
		}
		if err := subtype(ts[i].Type, u.Type, gamma); err != nil {
			return fmt.Errorf("parameter %d: %w", i, err)
		}
	}
	return nil
}

func sameAnnotations(a, b *FunctionType) bool {
	if len(a.Annotations) != len(b.Annotations) {
		return false
	}
	x := slices.Clone(a.Annotations)
	y := slices.Clone(b.Annotations)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

func acceptsNull(t Type) bool {
	switch unwrapRec(t).(type) {
	case *NullType, *ReservedType, *OptionalType:
		return true
	}
	return false
}

func unwrapRec(t Type) Type {
	for {
		r, ok := t.(*RecursiveType)
		if !ok || r.inner == nil {
			return t
		}
		t = r.inner
	}
}

func isNat(t Type) bool {
	n, ok := t.(*NatType)
	return ok && n.size == 0
}

func isInt(t Type) bool {
	n, ok := t.(*IntType)
	return ok && n.size == 0
}

func notSubtype(t, u Type) error {
	return fmt.Errorf("%s is not a subtype of %s", t, u)
}

// LabelHash normalises a record or variant label to its wire form: the wire
// carries the hash, a textual type carries the name.
func LabelHash(name string) string {
	if _, ok := new(big.Int).SetString(name, 10); ok {
		return name
	}
	return HashString(name)
}

// subtypeService: methods are like record fields, and one the reader does not
// expect is simply unused.
func subtypeService(t Type, u *ServiceType, gamma []typePair) error {
	s, ok := t.(*ServiceType)
	if !ok {
		return notSubtype(t, u)
	}
	have := make(map[string]*FunctionType, len(s.Methods))
	for _, m := range s.Methods {
		have[m.Name] = m.Func
	}
	for _, m := range u.Methods {
		f, ok := have[m.Name]
		if !ok {
			return fmt.Errorf("method %s: missing", m.Name)
		}
		// A decoded service can carry an unresolved signature.
		if f == nil || m.Func == nil {
			return fmt.Errorf("method %s: unresolved signature", m.Name)
		}
		if err := subtype(f, m.Func, gamma); err != nil {
			return fmt.Errorf("method %s: %w", m.Name, err)
		}
	}
	return nil
}
