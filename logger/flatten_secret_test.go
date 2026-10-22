package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenSecretTag(t *testing.T) {
	cfg := struct { //nolint:gosec // fake credentials to test masking
		Hidden       string `secret:"true"`
		HiddenEmpty  string `secret:"true"`
		HiddenInt    int    `secret:"true"`
		HiddenNilPtr *int   `secret:"true"`
		RevealedKey  string `secret:"false"`
		RevealedURL  string `secret:"false"`
		Untouched    string `secret:"maybe"`
	}{
		Hidden:      "q1w2e3",
		HiddenInt:   42,
		RevealedKey: "partition-1",
		RevealedURL: "postgres://user:pass@host/db",
		Untouched:   "value",
	}

	assert.Equal(t, map[string]any{ //nolint:gosec // fake credentials to test masking
		"Hidden":      maskedValue,
		"HiddenEmpty": "",
		"HiddenInt":   maskedValue,
		"RevealedKey": "partition-1",
		"RevealedURL": "postgres://user:pass@host/db",
		"Untouched":   "value",
	}, flattenPrefixed(cfg, ""))
}

func TestFlattenSensitiveFieldNames(t *testing.T) {
	str := "abc"
	cfg := struct {
		Key                string
		APIKey             string
		apikey             string // unexported: never printed
		UnleashToken       string
		DBPassword         string
		KafkaPass          string
		Pwd                string
		ClientSecret       string
		PostgresDSN        string
		Salt               string
		Credentials        string
		Auth               string
		ProxyAuthorization string
		TLSCert            string
		PrivateKeyPEM      string
		CertFile           string
		ServiceCredential  string
		EmptyToken         string
		PtrToken           *string
		Tokens             []string
		APIKeys            []string
		Keywords           string
		MaxTokens          int
		PartitionKey       bool
	}{
		Key:                "k",
		APIKey:             "k",
		apikey:             "k",
		UnleashToken:       "t",
		DBPassword:         "p",
		KafkaPass:          "p",
		Pwd:                "p",
		ClientSecret:       "s",
		PostgresDSN:        "host=db password=p",
		Salt:               "s",
		Credentials:        "{}",
		ServiceCredential:  "{}",
		Auth:               "basic",
		ProxyAuthorization: "Bearer t",
		TLSCert:            "-----BEGIN CERTIFICATE-----",
		PrivateKeyPEM:      "-----BEGIN PRIVATE KEY-----",
		CertFile:           "/etc/tls/tls.crt",
		PtrToken:           &str,
		Tokens:             []string{"a", "b"},
		APIKeys:            []string{"a"},
		Keywords:           "foo,bar",
		MaxTokens:          10,
		PartitionKey:       true,
	}

	assert.Equal(t, map[string]any{
		"Key":                maskedValue,
		"APIKey":             maskedValue,
		"UnleashToken":       maskedValue,
		"DBPassword":         maskedValue,
		"KafkaPass":          maskedValue,
		"Pwd":                maskedValue,
		"ClientSecret":       maskedValue,
		"PostgresDSN":        maskedValue,
		"Salt":               maskedValue,
		"Credentials":        maskedValue,
		"ServiceCredential":  maskedValue,
		"Auth":               maskedValue,
		"ProxyAuthorization": maskedValue,
		"TLSCert":            maskedValue,
		"PrivateKeyPEM":      maskedValue,
		"CertFile":           "/etc/tls/tls.crt",
		"EmptyToken":         "",
		"PtrToken":           maskedValue,
		"Tokens":             maskedValue,
		"APIKeys":            maskedValue,
		"Keywords":           "foo,bar",
		"MaxTokens":          10,
		"PartitionKey":       true,
	}, flattenPrefixed(cfg, ""))
}

func TestFlattenURLRedaction(t *testing.T) {
	const cacheURL = "redis://:pw@cache:6379"

	cfg := struct { //nolint:gosec // fake credentials to test redaction
		DB        string
		Webhook   string
		Plain     string
		NoCreds   string
		Broken    string
		URLs      []string
		Endpoints []string `secret:"false"`
	}{
		DB:        "postgres://user:s3cr3t@db:5432/app?sslmode=disable",
		Webhook:   "https://hooks.example.com/notify?channel=ops&token=abc&API_KEY=def&b=1",
		Plain:     "not a url",
		NoCreds:   "https://user@example.com/path?q=1",
		Broken:    "http://user:p@ss@[::1",
		URLs:      []string{cacheURL, "http://localhost:8080"},
		Endpoints: []string{cacheURL},
	}

	assert.Equal(t, map[string]any{ //nolint:gosec // redacted fake credentials
		"DB":        "postgres://user:***@db:5432/app?sslmode=disable",
		"Webhook":   "https://hooks.example.com/notify?channel=ops&token=***&API_KEY=***&b=1",
		"Plain":     "not a url",
		"NoCreds":   "https://user@example.com/path?q=1",
		"Broken":    maskedValue,
		"URLs":      []string{"redis://:***@cache:6379", "http://localhost:8080"},
		"Endpoints": []string{cacheURL},
	}, flattenPrefixed(cfg, ""))
}

func TestFlattenSecretTagPropagation(t *testing.T) {
	type db struct {
		Host     string
		Port     int
		Password string
		Debug    string `secret:"false"`
	}
	cfg := struct {
		Hidden   db `secret:"true"`
		Revealed db `secret:"false"`
		Auto     db
		Map      map[string]string
	}{
		Hidden:   db{Host: "h", Port: 1, Password: "p", Debug: "d"},
		Revealed: db{Host: "h", Port: 1, Password: "p"},
		Auto:     db{Host: "h", Port: 1, Password: "p"},
		Map:      map[string]string{"api_key": "k", "region": "us", "Authorization": "Bearer t"},
	}

	assert.Equal(t, map[string]any{
		"Hidden.Host":       maskedValue,
		"Hidden.Port":       maskedValue,
		"Hidden.Password":   maskedValue,
		"Hidden.Debug":      "d",
		"Revealed.Host":     "h",
		"Revealed.Port":     1,
		"Revealed.Password": "p",
		"Revealed.Debug":    "",
		"Auto.Host":         "h",
		"Auto.Port":         1,
		"Auto.Password":     maskedValue,
		"Auto.Debug":        "",
		"Map.api_key":       maskedValue,
		"Map.region":        "us",
		"Map.Authorization": maskedValue,
	}, flattenPrefixed(cfg, ""))
}
