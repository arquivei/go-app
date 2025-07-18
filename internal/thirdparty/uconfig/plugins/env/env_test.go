package env_test

import (
	"os"
	"testing"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/internal/f"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvBasic(t *testing.T) {
	envs := map[string]string{
		"GOHARD":               "T",
		"VERSION":              "0.2",
		"REDIS_HOST":           "redis-host",
		"REDIS_PORT":           "6379",
		"RETHINK_HOST_ADDRESS": "rethink-cluster",
		"RETHINK_HOST_PORT":    "28015",
		"RETHINK_DB":           "",
	}

	expect := f.Config{
		Anon: f.Anon{
			Version: "0.2",
		},

		GoHard: true,

		Redis: f.Redis{
			Host: "redis-host",
			Port: 6379,
		},

		Rethink: f.RethinkConfig{
			Host: f.Host{
				Address: "rethink-cluster",
				Port:    "28015",
			},
			DB: "",
		},
	}

	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(t, err, "failed to set env var %s", key)
	}

	value := f.Config{Rethink: f.RethinkConfig{DB: "must-be-override-by-empty-env"}}

	conf, err := uconfig.New(&value, env.New())
	require.NoError(t, err)

	err = conf.Parse()
	require.NoError(t, err)

	assert.Equal(t, expect, value)
}

type fEnv struct {
	Address string `env:"MY_HOST_NAME"`
}

func TestEnvTag(t *testing.T) {
	envs := map[string]string{
		"MY_HOST_NAME": "https://blah.bleh",
	}

	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(t, err, "failed to set env var %s", key)
	}

	expect := fEnv{
		Address: "https://blah.bleh",
	}

	value := fEnv{}

	conf, err := uconfig.New(&value, env.New())
	if err != nil {
		t.Fatal(err)
	}

	err = conf.Parse()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expect, value)
}
