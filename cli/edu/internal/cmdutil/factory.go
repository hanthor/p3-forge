// Package cmdutil holds the Factory injected into every command package.
package cmdutil

import (
	"context"
	"io"
	"os"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/config"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/contexts"
	"github.com/LF-Certification/p3-forge/cli/edu/sdk/output"
)

// IOStreams are the process streams. Tests replace them with buffers.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// Factory carries shared state and lazily resolved dependencies. Command
// constructors take *Factory and read it at run time, not at build time,
// so persistent flags parsed by Cobra are visible.
type Factory struct {
	IOStreams IOStreams
	// Version is the host semver (internal/build.Version).
	Version string

	// Root persistent flags. Cobra writes them before any RunE runs.
	ContextFlag     string // -c/--context
	OutputFlag      string // -o/--output
	Yes             bool   // --yes
	InsecureStorage bool   // --insecure-storage

	// Config loads $XDG_CONFIG_HOME/edu/config.yaml.
	Config func() (*config.Config, error)
	// Context resolves the active context: -c, EDU_CONTEXT, config file.
	Context func() (contexts.Context, error)
	// Token returns a valid access token for the active context. It
	// refreshes when the cached token expires within 60 seconds. The auth
	// package installs it in main.
	Token func(ctx context.Context) (string, error)
	// Executable returns the absolute path of the running edu binary. The
	// plugin manager exports it as EDU_REEXEC.
	Executable func() (string, error)
}

// New builds a Factory bound to the process streams.
func New(version string) *Factory {
	f := &Factory{
		IOStreams: IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		Version:   version,
		Config:    config.Load,
		Executable: func() (string, error) {
			return os.Executable()
		},
	}
	f.Context = func() (contexts.Context, error) {
		cfg, err := f.Config()
		if err != nil {
			return contexts.Context{}, err
		}
		return contexts.Resolve(f.ContextFlag, os.Getenv("EDU_CONTEXT"), cfg.CurrentContext)
	}
	return f
}

// Output returns the validated -o mode.
func (f *Factory) Output() (output.Mode, error) {
	return output.ParseMode(f.OutputFlag)
}
