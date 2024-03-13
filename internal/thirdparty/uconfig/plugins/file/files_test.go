package file_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/internal/f"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/file"
	"github.com/stretchr/testify/assert"
)

func TestFiles(t *testing.T) {
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
			Db: "base",
		},
	}

	files := file.Files{
		{"testdata/config_rethink.json", json.Unmarshal, true},
		{"testdata/config_partial.json", json.Unmarshal, true},
	}

	value := f.Config{}

	os.Args = os.Args[:1]
	_, err := uconfig.Classic(&value, files)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expect, value)
}
