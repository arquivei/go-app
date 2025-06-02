package uconfig_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/internal/f"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/file"
)

func TestMust(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Must should not panic, but did: %v", r)
		}
	}()

	value := f.Config{}
	uconfig.Must(&value)
}

func TestMustPanic(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Error("Was expecting panic but got nil")
		}

		expectErr := "read testdata/classic.json: file already closed"

		if err, ok := r.(error); !ok || err.Error() != expectErr {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	open, err := os.Open("testdata/classic.json")
	if err != nil {
		t.Fatal(err)
	}

	// nolint: errcheck, gosec
	open.Close() // close it so we get an error!
	reader := file.NewReader(open, json.Unmarshal)

	var buf bytes.Buffer
	uconfig.UsageOutput = &buf

	value := f.Config{}
	uconfig.Must(&value, reader)
}

func TestMustPanicNew(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Error("Was expecting panic but got nil")
		}

		expectErr := "unexpected type, expecting a pointer to struct"

		if err, ok := r.(error); !ok || err.Error() != expectErr {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	conf := f.Config{}

	// passing non-pointer.
	uconfig.Must(conf)
}
