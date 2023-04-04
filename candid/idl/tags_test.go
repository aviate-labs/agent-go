package idl

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseTags(t *testing.T) {
	t.Run("custom struct", func(t *testing.T) {
		type custom struct {
			Name string
		}
		field := reflect.TypeOf(custom{}).Field(0)
		tag := parseTags(field)
		if tag.name != "name" {
			t.Errorf("got %q, want %q", tag.name, "name")
		}
	})

	t.Run("custom struct", func(t *testing.T) {
		type custom struct {
			Name string `ic:"anotherName"`
		}
		field := reflect.TypeOf(custom{}).Field(0)
		tag := parseTags(field)
		if tag.name != "anotherName" {
			t.Errorf("got %q, want %q", tag.name, "name")
		}
	})

	type test struct {
		name string
		ic   string
	}
	tests := []test{
		{
			name: "empty",
			ic:   "",
		},
		{
			name: "name only",
			ic:   "name",
		},
		{
			name: "name and options",
			ic:   "name,option1,option2",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: "Name",
				Tag:  reflect.StructTag("ic:" + test.ic),
				Type: reflect.TypeOf(""),
			}
			tag := parseTags(field)
			if test.ic != "" && tag.name != strings.Split(test.ic, ",")[0] {
				t.Errorf("got %q, want %q", tag.name, test.ic)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	type test struct {
		name string
		in   any
		want map[string]any
		err  string
	}
	tests := []test{
		{
			name: "empty struct",
			in:   struct{}{},
			want: map[string]any{},
		},
		{
			name: "struct with one field",
			in: struct {
				Name string
			}{},
			want: map[string]any{
				"name": "",
			},
		},
		{
			name: "struct with custom name",
			in: struct {
				Name string `ic:"anotherName"`
			}{
				Name: "test",
			},
			want: map[string]any{
				"anotherName": "test",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := StructToMap(test.in)
			if err != nil {
				if test.err == "" {
					t.Errorf("structToMap() error = %v, wantErr %v", err, test.err)
				} else if !strings.Contains(err.Error(), test.err) {
					t.Errorf("structToMap() error = %v, wantErr %v", err, test.err)
				}
				return
			}
			if test.err != "" {
				t.Errorf("structToMap() error = %v, wantErr %v", err, test.err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("structToMap() = %v, want %v", got, test.want)
			}
		})
	}
}
