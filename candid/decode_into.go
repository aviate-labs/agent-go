package candid

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/niccolofant/agent-go/candid/idl"
)

var rawMessagePtrType = reflect.TypeOf((*idl.RawMessage)(nil))
var structFieldIndexes sync.Map // map[reflect.Type]map[string]int

type decodeIntoVisit struct {
	typ uintptr
	dst reflect.Type
}

func canUnmarshalDirect(t idl.Type, v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return false
	}
	dst := rv.Elem()
	if !hasUsefulSkip(t, dst, false, make(map[decodeIntoVisit]bool), make(map[uintptr]bool)) {
		return false
	}
	return canDecodeIntoValue(t, dst, make(map[decodeIntoVisit]bool), make(map[uintptr]bool))
}

func unmarshalDirect(t idl.Type, r *bytes.Reader, v any) error {
	return decodeIntoValue(t, r, reflect.ValueOf(v).Elem())
}

func canDecodeIntoValue(t idl.Type, dst reflect.Value, seen map[decodeIntoVisit]bool, skipSeen map[uintptr]bool) bool {
	if !dst.IsValid() {
		return false
	}
	if dst.CanAddr() && dst.Addr().Type() == rawMessagePtrType {
		return true
	}

	if key, ok := decodeIntoVisitKey(t, dst); ok {
		if seen[key] {
			return true
		}
		seen[key] = true
		defer delete(seen, key)
	}

	switch t := t.(type) {
	case *idl.RecordType:
		if dst.Kind() != reflect.Struct {
			return false
		}
		for _, f := range t.Fields {
			field, ok := fieldByCandidName(dst, f.Name)
			if !ok {
				if !canSkipValue(f.Type, skipSeen) {
					return false
				}
				continue
			}
			if !canDecodeIntoValue(f.Type, field, seen, skipSeen) {
				return false
			}
		}
		return true
	case *idl.VectorType:
		if dst.Kind() != reflect.Slice && dst.Kind() != reflect.Array {
			return false
		}
		return canDecodeIntoValue(t.Type, reflect.New(dst.Type().Elem()).Elem(), seen, skipSeen)
	case *idl.OptionalType:
		if dst.Kind() != reflect.Pointer {
			return false
		}
		return canDecodeIntoValue(t.Type, reflect.New(dst.Type().Elem()).Elem(), seen, skipSeen)
	case *idl.VariantType:
		if dst.Kind() != reflect.Struct {
			return false
		}
		for _, f := range t.Fields {
			field, ok := fieldByCandidName(dst, f.Name)
			if !ok {
				if !canSkipValue(f.Type, skipSeen) {
					return false
				}
				continue
			}
			if field.Kind() != reflect.Pointer {
				return false
			}
			if !canDecodeIntoValue(f.Type, reflect.New(field.Type().Elem()).Elem(), seen, skipSeen) {
				return false
			}
		}
		return true
	case *idl.EmptyType, *idl.ReservedType, *idl.Service, idl.Service:
		return false
	default:
		return canDecodeScalarInto(t, dst)
	}
}

func canSkipValue(t idl.Type, seen map[uintptr]bool) bool {
	if key, ok := typePointerKey(t); ok {
		if seen[key] {
			return true
		}
		seen[key] = true
		defer delete(seen, key)
	}

	switch t := t.(type) {
	case *idl.RecordType:
		for _, f := range t.Fields {
			if !canSkipValue(f.Type, seen) {
				return false
			}
		}
		return true
	case *idl.VectorType:
		return canSkipValue(t.Type, seen)
	case *idl.OptionalType:
		return canSkipValue(t.Type, seen)
	case *idl.VariantType:
		for _, f := range t.Fields {
			if !canSkipValue(f.Type, seen) {
				return false
			}
		}
		return true
	case *idl.EmptyType:
		return false
	case *idl.NullType, *idl.ReservedType, *idl.BoolType, *idl.NatType,
		*idl.IntType, *idl.FloatType, *idl.TextType, *idl.PrincipalType,
		*idl.FunctionType, *idl.Service, idl.Service, *idl.FutureType:
		return true
	default:
		return false
	}
}

func skipIsUseful(t idl.Type, seen map[uintptr]bool) bool {
	if key, ok := typePointerKey(t); ok {
		if seen[key] {
			return false
		}
		seen[key] = true
		defer delete(seen, key)
	}
	switch t := t.(type) {
	case *idl.RecordType, *idl.VectorType, *idl.VariantType:
		return true
	case *idl.OptionalType:
		return skipIsUseful(t.Type, seen)
	default:
		return false
	}
}

func hasUsefulSkip(t idl.Type, dst reflect.Value, repeated bool, seen map[decodeIntoVisit]bool, skipSeen map[uintptr]bool) bool {
	if !dst.IsValid() {
		return false
	}
	if dst.CanAddr() && dst.Addr().Type() == rawMessagePtrType {
		return false
	}
	if key, ok := decodeIntoVisitKey(t, dst); ok {
		if seen[key] {
			return false
		}
		seen[key] = true
		defer delete(seen, key)
	}

	switch t := t.(type) {
	case *idl.RecordType:
		if dst.Kind() != reflect.Struct {
			return false
		}
		for _, f := range t.Fields {
			field, ok := fieldByCandidName(dst, f.Name)
			if !ok {
				if skipIsUseful(f.Type, skipSeen) || (repeated && canSkipValue(f.Type, skipSeen)) {
					return true
				}
				continue
			}
			if hasUsefulSkip(f.Type, field, repeated, seen, skipSeen) {
				return true
			}
		}
	case *idl.VectorType:
		if dst.Kind() != reflect.Slice && dst.Kind() != reflect.Array {
			return false
		}
		return hasUsefulSkip(t.Type, reflect.New(dst.Type().Elem()).Elem(), true, seen, skipSeen)
	case *idl.OptionalType:
		if dst.Kind() != reflect.Pointer {
			return false
		}
		return hasUsefulSkip(t.Type, reflect.New(dst.Type().Elem()).Elem(), repeated, seen, skipSeen)
	case *idl.VariantType:
		if dst.Kind() != reflect.Struct {
			return false
		}
		for _, f := range t.Fields {
			field, ok := fieldByCandidName(dst, f.Name)
			if !ok {
				continue
			}
			if field.Kind() != reflect.Pointer {
				continue
			}
			if hasUsefulSkip(f.Type, reflect.New(field.Type().Elem()).Elem(), repeated, seen, skipSeen) {
				return true
			}
		}
	}
	return false
}

func decodeIntoValue(t idl.Type, r *bytes.Reader, dst reflect.Value) error {
	if dst.CanAddr() && dst.Addr().Type() == rawMessagePtrType {
		raw, err := t.Read(r)
		if err != nil {
			return err
		}
		dst.SetBytes(raw)
		return nil
	}

	switch t := t.(type) {
	case *idl.RecordType:
		if dst.Kind() != reflect.Struct {
			raw, err := t.Decode(r)
			if err != nil {
				return err
			}
			return idl.UnmarshalGo(t, raw, dst.Addr().Interface())
		}
		for _, f := range t.Fields {
			field, ok := fieldByCandidName(dst, f.Name)
			if !ok {
				if err := skipValue(f.Type, r); err != nil {
					return err
				}
				continue
			}
			if err := decodeIntoValue(f.Type, r, field); err != nil {
				return err
			}
		}
		return nil
	case *idl.VectorType:
		return decodeVectorInto(t, r, dst)
	case *idl.OptionalType:
		return decodeOptionalInto(t, r, dst)
	case *idl.VariantType:
		return decodeVariantInto(t, r, dst)
	default:
		raw, err := t.Decode(r)
		if err != nil {
			return err
		}
		return idl.UnmarshalGo(t, raw, dst.Addr().Interface())
	}
}

func decodeVectorInto(t *idl.VectorType, r *bytes.Reader, dst reflect.Value) error {
	n, err := decodeIntoLen(r)
	if err != nil {
		return err
	}
	switch dst.Kind() {
	case reflect.Slice:
		if dst.IsNil() || dst.Cap() < n {
			dst.Set(reflect.MakeSlice(dst.Type(), n, n))
		} else {
			dst.Set(dst.Slice(0, n))
		}
	case reflect.Array:
		if dst.Len() != n {
			return idl.NewUnmarshalGoError(nil, dst.Addr().Interface())
		}
	default:
		return idl.NewUnmarshalGoError(nil, dst.Addr().Interface())
	}
	for i := 0; i < n; i++ {
		if err := decodeIntoValue(t.Type, r, dst.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func decodeOptionalInto(t *idl.OptionalType, r *bytes.Reader, dst reflect.Value) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch b {
	case 0x00:
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	case 0x01:
		if dst.Kind() != reflect.Pointer {
			return idl.NewUnmarshalGoError(nil, dst.Addr().Interface())
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return decodeIntoValue(t.Type, r, dst.Elem())
	default:
		return fmt.Errorf("invalid option value: %x", b)
	}
}

func decodeVariantInto(t *idl.VariantType, r *bytes.Reader, dst reflect.Value) error {
	index, err := readULEB128Uint64(r)
	if err != nil {
		return err
	}
	if index >= uint64(len(t.Fields)) {
		return fmt.Errorf("invalid variant index: %v", index)
	}
	f := t.Fields[int(index)]
	field, ok := fieldByCandidName(dst, f.Name)
	if !ok {
		if err := skipValue(f.Type, r); err != nil {
			return err
		}
		return idl.NewUnmarshalGoError(nil, dst.Addr().Interface())
	}
	if field.Kind() != reflect.Pointer {
		return idl.NewUnmarshalGoError(nil, dst.Addr().Interface())
	}
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	return decodeIntoValue(f.Type, r, field.Elem())
}

func skipValue(t idl.Type, r *bytes.Reader) error {
	switch t := t.(type) {
	case *idl.RecordType:
		for _, f := range t.Fields {
			if err := skipValue(f.Type, r); err != nil {
				return err
			}
		}
		return nil
	case *idl.VectorType:
		n, err := decodeIntoLen(r)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			if err := skipValue(t.Type, r); err != nil {
				return err
			}
		}
		return nil
	case *idl.OptionalType:
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		switch b {
		case 0x00:
			return nil
		case 0x01:
			return skipValue(t.Type, r)
		default:
			return fmt.Errorf("invalid option value: %x", b)
		}
	case *idl.VariantType:
		index, err := readULEB128Uint64(r)
		if err != nil {
			return err
		}
		if index >= uint64(len(t.Fields)) {
			return fmt.Errorf("invalid variant index: %v", index)
		}
		return skipValue(t.Fields[int(index)].Type, r)
	case *idl.NullType, *idl.ReservedType:
		return nil
	case *idl.EmptyType:
		return fmt.Errorf("cannot skip empty type")
	case *idl.BoolType:
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b != 0x00 && b != 0x01 {
			return fmt.Errorf("invalid bool values: %x", b)
		}
		return nil
	case *idl.NatType:
		if t.Base() == 0 {
			return skipLEB128(r)
		}
		return skipBytes(r, int(t.Base()))
	case *idl.IntType:
		if t.Base() == 0 {
			return skipLEB128(r)
		}
		return skipBytes(r, int(t.Base()))
	case *idl.FloatType:
		return skipBytes(r, int(t.Base()))
	case *idl.TextType:
		n, err := decodeIntoLen(r)
		if err != nil {
			return err
		}
		return skipBytes(r, n)
	case *idl.PrincipalType:
		return skipPrincipal(r)
	case *idl.FunctionType:
		return skipFunction(r)
	case *idl.Service, idl.Service:
		return skipService(r)
	case *idl.FutureType:
		return skipFuture(r)
	default:
		return fmt.Errorf("cannot skip unsupported type %T", t)
	}
}

func canDecodeScalarInto(t idl.Type, dst reflect.Value) bool {
	if !dst.CanAddr() {
		return false
	}
	switch t.(type) {
	case *idl.NullType, *idl.BoolType, *idl.NatType, *idl.IntType,
		*idl.FloatType, *idl.TextType, *idl.PrincipalType, *idl.FunctionType,
		*idl.FutureType:
		return true
	default:
		return false
	}
}

func skipFunction(r *bytes.Reader) error {
	b0, err := r.ReadByte()
	if err != nil {
		return err
	}
	b1, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b0 != 0x01 || b1 != 0x01 {
		return fmt.Errorf("invalid func reference: [%d %d]", b0, b1)
	}
	n, err := decodeIntoLen(r)
	if err != nil {
		return err
	}
	if err := skipBytes(r, n); err != nil {
		return err
	}
	n, err = decodeIntoLen(r)
	if err != nil {
		return err
	}
	return skipBytes(r, n)
}

func skipService(r *bytes.Reader) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b != 0x01 {
		return fmt.Errorf("invalid service reference: %d", b)
	}
	n, err := decodeIntoLen(r)
	if err != nil {
		return err
	}
	return skipBytes(r, n)
}

func skipPrincipal(r *bytes.Reader) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b != 0x01 {
		return fmt.Errorf("cannot decode principal")
	}
	n, err := decodeIntoLen(r)
	if err != nil {
		return err
	}
	return skipBytes(r, n)
}

func skipFuture(r *bytes.Reader) error {
	m, err := decodeIntoLen(r)
	if err != nil {
		return err
	}
	if err := skipLEB128(r); err != nil {
		return err
	}
	return skipBytes(r, m)
}

func decodeIntoLen(r *bytes.Reader) (int, error) {
	l, err := readULEB128Uint64(r)
	if err != nil {
		return 0, err
	}
	if l > uint64(r.Len()) || l > uint64(int(^uint(0)>>1)) {
		return 0, fmt.Errorf("invalid length %d with %d bytes remaining", l, r.Len())
	}
	return int(l), nil
}

func skipLEB128(r *bytes.Reader) error {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b < 0x80 {
			return nil
		}
	}
}

func readULEB128Uint64(r *bytes.Reader) (uint64, error) {
	var value uint64
	var shift uint
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if shift == 63 && b > 1 {
			return 0, fmt.Errorf("uleb128 overflows uint64")
		}
		value |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return value, nil
		}
		shift += 7
		if shift >= 64 || i >= 9 {
			return 0, fmt.Errorf("uleb128 overflows uint64")
		}
	}
}

func skipBytes(r *bytes.Reader, n int) error {
	if n < 0 || n > r.Len() {
		return io.ErrUnexpectedEOF
	}
	_, err := r.Seek(int64(n), io.SeekCurrent)
	return err
}

func fieldByCandidName(dst reflect.Value, name string) (reflect.Value, bool) {
	name = lowerFirstCharacter(name)
	indexes := cachedStructFieldIndexes(dst.Type())
	i, ok := indexes[name]
	if !ok {
		return reflect.Value{}, false
	}
	field := dst.Field(i)
	if !field.CanSet() {
		return reflect.Value{}, false
	}
	return field, true
}

func cachedStructFieldIndexes(t reflect.Type) map[string]int {
	if indexes, ok := structFieldIndexes.Load(t); ok {
		return indexes.(map[string]int)
	}
	indexes := make(map[string]int, t.NumField()*2)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := idl.ParseTags(field)
		if _, ok := indexes[tag.Name]; !ok {
			indexes[tag.Name] = i
		}
		hash := idl.HashString(tag.Name)
		if _, ok := indexes[hash]; !ok {
			indexes[hash] = i
		}
	}
	actual, _ := structFieldIndexes.LoadOrStore(t, indexes)
	return actual.(map[string]int)
}

func lowerFirstCharacter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func decodeIntoVisitKey(t idl.Type, dst reflect.Value) (decodeIntoVisit, bool) {
	key, ok := typePointerKey(t)
	if !ok {
		return decodeIntoVisit{}, false
	}
	return decodeIntoVisit{typ: key, dst: dst.Type()}, true
}

func typePointerKey(t idl.Type) (uintptr, bool) {
	rv := reflect.ValueOf(t)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer {
		return 0, false
	}
	return rv.Pointer(), true
}
