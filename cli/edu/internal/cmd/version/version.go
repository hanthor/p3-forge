// Package version implements `edu version`.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
)

// NewCmd builds `edu version`.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the edu version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(f.IOStreams.Out, "edu %s\n", f.Version)
			return err
		},
	}
}
