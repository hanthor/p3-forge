// Package plugin implements `edu plugin install|list|remove`.
package plugin

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/pluginmgr"
	"github.com/LF-Certification/p3-forge/cli/edu/sdk/output"
	sdkplugin "github.com/LF-Certification/p3-forge/cli/edu/sdk/plugin"
)

// NewCmd builds the `edu plugin` group.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Install, list and remove plugins",
	}
	cmd.AddCommand(newInstallCmd(f), newListCmd(f), newRemoveCmd(f))
	return cmd
}

func newInstallCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "install <path>",
		Short: "Install a plugin from a binary or a Go main package directory",
		Long: `Install a plugin from a local path.

<path> is an executable file, or a directory holding a Go main package
directly or under cmd/edu-*. A directory is built with "go build". The
host runs "<binary> __metadata", validates the manifest and stores both
under the plugin directory. An existing plugin of the same name is
replaced.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := pluginmgr.Install(cmd.Context(), f, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "installed %s %s to %s\n", p.Manifest.Name, p.Manifest.Version, p.Dir)
			return nil
		},
	}
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := f.Output()
			if err != nil {
				return err
			}
			installed, err := pluginmgr.List()
			if err != nil {
				return err
			}
			rows := make([]sdkplugin.Manifest, 0, len(installed))
			for _, p := range installed {
				rows = append(rows, p.Manifest)
			}
			return output.Print(f.IOStreams.Out, mode, rows, []output.Column[sdkplugin.Manifest]{
				{Header: "name", Value: func(m sdkplugin.Manifest) string { return m.Name }},
				{Header: "version", Value: func(m sdkplugin.Manifest) string { return m.Version }},
				{Header: "summary", Value: func(m sdkplugin.Manifest) string { return m.Summary }},
			})
		},
	}
}

func newRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return Names(), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pluginmgr.Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "removed %s\n", args[0])
			return nil
		},
	}
}

// Names returns "name\tsummary" completions for installed plugins. Hidden
// plugins are omitted. Errors yield no completions.
func Names() []string {
	installed, err := pluginmgr.List()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(installed))
	for _, p := range installed {
		if p.Manifest.Hidden {
			continue
		}
		names = append(names, p.Manifest.Name+"\t"+p.Manifest.Summary)
	}
	return names
}
