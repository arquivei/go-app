package uconfig_test

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/internal/f"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// for go test framework.
	flag.Parse()

	os.Exit(m.Run())
}

func TestClassicBasic(t *testing.T) {
	expect := f.Config{
		Anon: f.Anon{
			Version: "from-flags",
		},

		GoHard: true,

		Redis: f.Redis{
			Host: "from-envs",
			Port: 6379,
		},

		Rethink: f.RethinkConfig{
			Host: f.Host{
				Address: "rethink-cluster",
				Port:    "28015",
			},
			DB: "base",
		},
	}

	files := uconfig.Files{
		{"testdata/classic.json", json.Unmarshal, true},
	}

	value := f.Config{}

	// set some env vars to test env var and plugin orders.
	err := os.Setenv("VERSION", "bad-value-overrided-with-flags")
	require.NoError(t, err)
	err = os.Setenv("REDIS_HOST", "from-envs")
	require.NoError(t, err)
	// patch the os.Args. for our tests.
	os.Args = append(os.Args[:1], "-version=from-flags")

	_, err = uconfig.Classic(&value, files)
	require.NoError(t, err)

	assert.Equal(t, expect, value)
}

func TestClassicBadPlugin(t *testing.T) {
	var badPlugin BadPlugin

	config := f.Config{}

	_, err := uconfig.Classic(&config, nil, badPlugin)

	if err == nil {
		t.Error("expected error for bad plugin, got nil")
	}

	if err.Error() != "unsupported plugins. Expecting a Walker or Visitor" {
		t.Errorf("Expected unsupported plugin error, got: %v", err)
	}
}
