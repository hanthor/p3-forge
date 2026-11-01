// Package root builds the `edu` root command and dispatches unknown
// commands to plugins. See Run for the dispatch mechanism.
package root

import (
	"github.com/spf13/cobra"

	authcmd "github.com/LF-Certification/p3-forge/cli/edu/internal/cmd/auth"
	contextcmd "github.com/LF-Certification/p3-forge/cli/edu/internal/cmd/context"
	plugincmd "github.com/LF-Certification/p3-forge/cli/edu/internal/cmd/plugin"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmd/version"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
)

// NewCmd builds the root command with its global flags and core commands.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "edu",
		Short:         "Education platform CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			_, err := f.Output()
			return err
		},
		ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return plugincmd.Names(), cobra.ShellCompDirectiveNoFileComp
		},
	}
	cmd.SetIn(f.IOStreams.In)
	cmd.SetOut(f.IOStreams.Out)
	cmd.SetErr(f.IOStreams.ErrOut)
	cmd.SetUsageTemplate(usageTemplate(cmd))

	pf := cmd.PersistentFlags()
	pf.StringVarP(&f.ContextFlag, "context", "c", "", "Context to use: dev, staging or prod (default from EDU_CONTEXT or the config file)")
	pf.StringVarP(&f.OutputFlag, "output", "o", "table", "Output format: table or json")
	pf.BoolVar(&f.Yes, "yes", false, "Skip confirmation prompts")
	pf.BoolVar(&f.InsecureStorage, "insecure-storage", false, "Store credentials in a plain file instead of the system keychain")

	cmd.AddCommand(
		version.NewCmd(f),
		contextcmd.NewCmd(f),
		authcmd.NewCmd(f),
		plugincmd.NewCmd(f),
	)
	return cmd
}
