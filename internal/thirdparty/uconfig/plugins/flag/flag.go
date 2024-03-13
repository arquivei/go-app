// Package flag provides flags support for uconfig
package flag

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/arquivei/go-app/internal/thirdparty/uconfig/flat"
	"github.com/arquivei/go-app/internal/thirdparty/uconfig/plugins"
)

const tag = "flag"

func init() {
	plugins.RegisterTag(tag)
}

// ErrorHandling defines how FlagSet.Parse behaves if the parse fails.
type ErrorHandling flag.ErrorHandling

// These constants cause FlagSet.Parse to behave as described if the parse fails.
const (
	ContinueOnError = ErrorHandling(flag.ContinueOnError)
	ExitOnError     = ErrorHandling(flag.ExitOnError)
	PanicOnError    = ErrorHandling(flag.PanicOnError)
)

// New returns a new Flags
func New(name string, errorHandling ErrorHandling, args []string) plugins.Plugin {
	fs := flag.NewFlagSet(name, flag.ErrorHandling(errorHandling))
	fs.Usage = func() {}

	return &visitor{
		fs:   fs,
		args: args,
	}
}

// Standard returns a set of flags configured in the common way.
// It is same as: `New(os.Args[0], ContinueOnError, os.Args[1:])`
func Standard() plugins.Plugin {
	return New(os.Args[0], ContinueOnError, os.Args[1:])
}

var _ plugins.Visitor = (*visitor)(nil)

type visitor struct {
	fs   *flag.FlagSet
	args []string
}

func (v *visitor) Parse() error {
	err := v.fs.Parse(v.args)

	if errors.Is(err, flag.ErrHelp) {
		return plugins.ErrUsage
	}

	return err
}

func (v *visitor) Visit(fields flat.Fields) error {
	for _, f := range fields {
		usage, _ := f.Tag("usage")

		name, ok := f.Tag(tag)
		if name == "-" {
			continue
		}

		if !ok || name == "" {
			name = f.Name()
			name = strings.Replace(name, ".", "-", -1)
			name = strings.ToLower(name)
		}

		f.Meta()[tag] = "-" + name
		v.fs.Var(f, name, usage)
	}

	return nil
}
