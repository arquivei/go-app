package uconfig

import (
	"errors"
	"os"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/defaults"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/env"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/file"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins/flag"
)

// Files represents a set of file paths and the appropriate unmarshaller function.
type Files = file.Files

// Classic creates a uconfig manager with defaults,environment variables,
// and flags (in that order) and optionally file loaders based on the provided
// Files map and parses them right away.
func Classic(conf interface{}, files Files, userPlugins ...plugins.Plugin) (Config, error) {
	fps := files.Plugins()

	ps := make([]plugins.Plugin, 0, len(fps)+3+len(userPlugins))

	// first defaults
	ps = append(ps, defaults.New())
	// then files
	ps = append(ps, fps...)
	// followed by env and flags
	ps = append(ps, env.New(), flag.Standard())
	// then any user pugins, often just _secret_.
	ps = append(ps, userPlugins...)

	c, err := New(conf, ps...)

	if err != nil {
		return c, err
	}

	err = c.Parse()
	if errors.Is(err, ErrUsage) {
		c.Usage()
		os.Exit(0)
	}

	return c, err
}
