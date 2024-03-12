package defaults_test

import (
	"testing"
	"time"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/defaults"
	"github.com/stretchr/testify/assert"
)

type fDefaults struct {
	Address string        `default:"https://blah.bleh"`
	Bases   []string      `default:"list,blah"`
	Timeout time.Duration `default:"5s"`
	Ignored string
}

func TestDefaultTag(t *testing.T) {

	expect := fDefaults{
		Address: "https://blah.bleh",
		Bases:   []string{"list", "blah"},
		Timeout: 5 * time.Second,
		Ignored: "not-empty",
	}

	value := fDefaults{Ignored: "not-empty"}

	conf, err := uconfig.New(&value, defaults.New())
	if err != nil {
		t.Fatal(err)
	}

	err = conf.Parse()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expect, value)
}
