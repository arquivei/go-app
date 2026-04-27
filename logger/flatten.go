package logger

import (
	"fmt"
	"reflect"
	"strings"
)

// Flatten transforms a struct into a flattened string, like: a.b.c: 'val', c.d: 'val'
// Pointer values will translate into memory addresses
//
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!! IMPORTANT SECURITY NOTE !!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//
// Flatten will place ALL keys of the given structure on
// the resulted string. This means that ANY sensitive data
// in the flattened structure WILL be exposed!
func Flatten(value any) string {
	m := flattenPrefixed(value, "")
	sb := strings.Builder{}
	for k, v := range m {
		sb.WriteString(k)
		sb.WriteString(": ")
		fmt.Fprintf(&sb, "'%v'", v)
		sb.WriteString(", ")
	}

	return strings.Trim(sb.String(), ", ")
}

func flattenPrefixed(value any, prefix string) map[string]any {
	m := make(map[string]any)
	flattenPrefixedToResult(value, prefix, m)
	return m
}

func flattenPrefixedToResult(value any, prefix string, m map[string]any) {
	if value == nil {
		return
	}

	base := ""
	if prefix != "" {
		base = prefix + "."
	}

	original := reflect.ValueOf(value)
	kind := original.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		original = reflect.Indirect(original)
		kind = original.Kind()
	}

	if original == (reflect.Value{}) {
		return
	}

	t := original.Type()

	switch kind {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			break
		}
		for _, childKey := range original.MapKeys() {
			childValue := original.MapIndex(childKey)
			if !childValue.CanInterface() {
				continue
			}
			flattenPrefixedToResult(childValue.Interface(), base+childKey.String(), m)
		}
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			isSecretStr, hasTag := t.Field(i).Tag.Lookup("secret")
			if hasTag && isSecretStr == "true" {
				continue
			}

			childValue := original.Field(i)
			if !childValue.CanInterface() {
				continue
			}
			childKey := t.Field(i).Name
			flattenPrefixedToResult(childValue.Interface(), base+childKey, m)
		}
	default:
		if prefix != "" {
			m[prefix] = value
		}
	}
}
