package plugin

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/LF-Certification/p3-forge/cli/edu/sdk/output"
)

// Serve runs root as an edu plugin and exits the process. It adds the
// hidden `__metadata` command, a persistent -o/--output flag defaulting to
// EDU_OUTPUT, and fills the manifest's schema version and command list
// from the Cobra tree when the caller left Commands nil. Errors print as
// `edu-<name>: <err>` and exit 1.
func Serve(root *cobra.Command, m Manifest) {
	os.Exit(run(root, m, os.Args[1:]))
}

// run is Serve without the process exit, for tests.
func run(root *cobra.Command, m Manifest, args []string) int {
	Prepare(root, m)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(root.ErrOrStderr(), "edu-%s: %v\n", m.Name, err)
		return 1
	}
	return 0
}

// Prepare wires root for the plugin contract without executing it. Serve
// calls it; tests call it directly and then root.Execute.
func Prepare(root *cobra.Command, m Manifest) {
	root.SilenceUsage = true
	root.SilenceErrors = true
	if root.Use == "" {
		root.Use = "edu-" + m.Name
	}
	mode, err := output.ParseMode(os.Getenv("EDU_OUTPUT"))
	if err != nil {
		mode = output.Table
	}
	if root.PersistentFlags().Lookup("output") == nil {
		root.PersistentFlags().StringP("output", "o", string(mode), "Output format: table or json")
	}
	root.AddCommand(newMetadataCmd(root, m))
}

func newMetadataCmd(root *cobra.Command, m Manifest) *cobra.Command {
	return &cobra.Command{
		Use:    "__metadata",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			m.SchemaVersion = ManifestSchemaVersion
			if m.Commands == nil {
				m.Commands = commandsOf(root)
			}
			return output.PrintJSON(cmd.OutOrStdout(), m)
		},
	}
}

// commandsOf lists runnable, visible leaf commands below root. Paths omit
// the root name so `edu help` can prefix the plugin name.
func commandsOf(root *cobra.Command) []Command {
	var out []Command
	var walk func(c *cobra.Command, path []string)
	walk = func(c *cobra.Command, path []string) {
		for _, sub := range c.Commands() {
			if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			p := append(append([]string(nil), path...), sub.Name())
			if sub.HasSubCommands() {
				walk(sub, p)
				continue
			}
			if sub.Runnable() {
				out = append(out, Command{Path: p, Summary: sub.Short})
			}
		}
	}
	walk(root, nil)
	if out == nil {
		out = []Command{}
	}
	return out
}

// Output returns the validated -o mode of cmd's tree.
func Output(cmd *cobra.Command) (output.Mode, error) {
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		return output.Table, errors.New("output flag not registered; call plugin.Serve")
	}
	return output.ParseMode(flag.Value.String())
}
