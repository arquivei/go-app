package logger

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// maskedValue replaces the value of a secret that is set, including secrets
// inside URLs.
const maskedValue = "***"

// sensitiveSuffixes are lowercase name suffixes that mark a field, map key or
// URL query parameter as secret.
var sensitiveSuffixes = []string{
	"key",
	"keys",
	"token",
	"tokens",
	"password",
	"passwords",
	"pass",
	"pwd",
	"secret",
	"secrets",
	"dsn",
	"salt",
	"credential",
	"credentials",
	"auth",
	"authorization",
	"cert",
	"pem",
}

// secretMode tells how the values under a field must be printed.
type secretMode int

const (
	// secretAuto masks values based on the field name and redacts URLs.
	secretAuto secretMode = iota
	// secretHide masks every value.
	secretHide
	// secretReveal prints every value as is.
	secretReveal
)

// Flatten transforms a struct into a flattened string, like: a.b.c: 'val', c.d: 'val'
// Pointer values will translate into memory addresses, except for strings.
//
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!! IMPORTANT SECURITY NOTE !!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//
// Flatten masks secrets on a best-effort basis:
//
//   - A field tagged with secret:"true" is printed as '***' (or empty when unset),
//     and so is every field nested under it.
//   - A field tagged with secret:"false" is printed as is, and so is every field
//     nested under it.
//   - Otherwise, a string field (string, *string or []string) whose name, or map
//     key, ends with a sensitive suffix is masked. The suffixes are Key, Token,
//     Password and Secret (and their plurals), Pass, Pwd, DSN, Salt, Credential,
//     Credentials, Auth, Authorization, Cert and PEM, case-insensitive.
//   - Otherwise, a string that contains "://" is parsed as a URL and has its
//     password and sensitive query parameters replaced by '***'. A string that
//     contains "://" but is not a valid URL is masked.
//
// The nearest secret tag wins. Secrets that do not match any of these rules,
// such as a token in a URL path, WILL be exposed unless tagged with secret:"true".
// See docs/secrets.md.
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
	flattenPrefixedToResult(value, prefix, "", secretAuto, m)
	return m
}

func flattenPrefixedToResult(value any, prefix, name string, mode secretMode, m map[string]any) {
	if value == nil {
		return
	}

	base := ""
	if prefix != "" {
		base = prefix + "."
	}

	original := reflect.ValueOf(value)
	kind := original.Kind()
	if kind == reflect.Pointer || kind == reflect.Interface {
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
			flattenPrefixedToResult(childValue.Interface(), base+childKey.String(), childKey.String(), mode, m)
		}
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			field := t.Field(i)
			childValue := original.Field(i)
			if !childValue.CanInterface() {
				continue
			}
			flattenPrefixedToResult(childValue.Interface(), base+field.Name, field.Name, fieldSecretMode(field, mode), m)
		}
	default:
		if prefix != "" {
			m[prefix] = leafValue(value, original, name, mode)
		}
	}
}

// fieldSecretMode returns the mode of a struct field: its own secret tag, if
// any, or the mode inherited from its parent.
func fieldSecretMode(field reflect.StructField, inherited secretMode) secretMode {
	switch field.Tag.Get("secret") {
	case "true":
		return secretHide
	case "false":
		return secretReveal
	default:
		return inherited
	}
}

// leafValue returns what is printed for a non-struct, non-map value.
func leafValue(value any, v reflect.Value, name string, mode secretMode) any {
	switch mode {
	case secretHide:
		return mask(!v.IsZero())
	case secretReveal:
		if v.Kind() == reflect.String {
			return v.String()
		}
		return value
	}

	switch {
	case v.Kind() == reflect.String:
		if isSensitiveName(name) {
			return mask(v.Len() > 0)
		}
		return redactURL(v.String())
	case v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String:
		if isSensitiveName(name) {
			return mask(v.Len() > 0)
		}
		redacted := make([]string, v.Len())
		for i := range redacted {
			redacted[i] = redactURL(v.Index(i).String())
		}
		return redacted
	default:
		return value
	}
}

func mask(isSet bool) string {
	if isSet {
		return maskedValue
	}
	return ""
}

func isSensitiveName(name string) bool {
	name = strings.ToLower(name)
	for _, suffix := range sensitiveSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// redactURL replaces the password and the sensitive query parameters of a URL
// with maskedValue. Strings that are not URLs are returned as is, and
// malformed URLs are masked, since they may still carry credentials.
func redactURL(s string) string {
	if !strings.Contains(s, "://") {
		return s
	}
	u, err := url.Parse(s)
	if err != nil {
		return maskedValue
	}
	u.RawQuery = redactQuery(u.RawQuery)
	if _, hasPassword := u.User.Password(); !hasPassword {
		return u.String()
	}
	// url.UserPassword would escape maskedValue, so the user info is written by hand.
	userInfo := url.User(u.User.Username()).String() + ":" + maskedValue + "@"
	u.User = nil
	return strings.Replace(u.String(), "://", "://"+userInfo, 1)
}

// redactQuery works on the raw query so the order and encoding of the
// parameters are preserved.
func redactQuery(rawQuery string) string {
	if rawQuery == "" {
		return rawQuery
	}
	params := strings.Split(rawQuery, "&")
	for i, param := range params {
		rawKey, _, hasValue := strings.Cut(param, "=")
		if !hasValue {
			continue
		}
		key, err := url.QueryUnescape(rawKey)
		if err != nil {
			key = rawKey
		}
		if isSensitiveName(key) {
			params[i] = rawKey + "=" + maskedValue
		}
	}
	return strings.Join(params, "&")
}
