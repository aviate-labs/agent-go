package idl

import (
	"reflect"
	"strings"
)

// lowerFirstCharacter returns a copy of the string with the first character in lower case.
// e.g. "UserName" -> "userName"
func lowerFirstCharacter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

type Tag struct {
	// Name is the name of the field in the struct.
	Name        string
	VariantType bool
}

func ParseTags(field reflect.StructField) Tag {
	icTag := field.Tag.Get("ic")
	if icTag == "" {
		return Tag{
			Name: lowerFirstCharacter(field.Name),
		}
	}
	var t Tag
	tags := strings.Split(icTag, ",")
	if len(tags) != 0 {
		t.Name = tags[0]
		for _, option := range tags[1:] {
			switch option {
			case "variant":
				t.VariantType = true
			default:
				// ignore unknown options
			}
		}
	}
	return t
}
