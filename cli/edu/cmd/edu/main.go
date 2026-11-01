// Command edu is the education platform CLI host.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/auth"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/build"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmd/root"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/pluginmgr"
)

func main() {
	f := cmdutil.New(build.Version)
	f.Token = auth.TokenResolver(f)
	if err := root.Run(context.Background(), f, os.Args[1:]); err != nil {
		// A plugin's exit code passes through; the plugin already reported.
		var exit *pluginmgr.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.Code)
		}
		fmt.Fprintf(f.IOStreams.ErrOut, "edu: %v\n", err)
		os.Exit(pluginmgr.ExitCode(err))
	}
}
