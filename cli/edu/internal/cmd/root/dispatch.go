package root

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/pluginmgr"
)

// Run executes argv for the edu host.
//
// It pre-scans args before Cobra sees them. Host flags before the first
// word are recognised by the root persistent flag set; the first word that
// is not a flag is the command. A core command, "help", "completion" or
// the Cobra completion hooks go to Cobra unchanged. Any other word is a
// plugin name: the remaining args are forwarded raw, after host flags in
// them (-c, -o, --yes, --insecure-storage) are also applied to the host.
// Cobra would otherwise strip or reject flags it does not own, so the
// plugin would not see its own flags.
func Run(ctx context.Context, f *cmdutil.Factory, args []string) error {
	cmd := NewCmd(f)
	d, ok := split(cmd, args)
	if !ok {
		cmd.SetArgs(args)
		return cmd.ExecuteContext(ctx)
	}
	if err := applyHostFlags(cmd.PersistentFlags(), d.rest); err != nil {
		return err
	}
	if d.complete {
		return pluginmgr.Complete(ctx, f, d.name, d.rest)
	}
	return pluginmgr.Run(ctx, f, d.name, d.rest)
}

// dispatch is a resolved plugin invocation.
type dispatch struct {
	name     string
	rest     []string
	complete bool
}

// split finds the plugin invocation in args. ok is false when Cobra must
// handle args: a core command, help, no command word, or a flag the host
// does not know. Host flags before the command word are applied.
func split(cmd *cobra.Command, args []string) (d dispatch, ok bool) {
	pf := cmd.PersistentFlags()
	i := 0
	for i < len(args) {
		a := args[i]
		if !strings.HasPrefix(a, "-") || a == "-" {
			break
		}
		fl, value, n, known := hostFlag(pf, args, i)
		if !known {
			return dispatch{}, false
		}
		if err := setFlag(fl, value); err != nil {
			return dispatch{}, false
		}
		i += n
	}
	if i >= len(args) {
		return dispatch{}, false
	}
	word := args[i]
	if word == cobra.ShellCompRequestCmd || word == cobra.ShellCompNoDescRequestCmd {
		// The last arg is the word being completed. A plugin name still
		// being typed is completed by Cobra from the root ValidArgsFunction.
		if i+2 >= len(args) || isCore(cmd, args[i+1]) {
			return dispatch{}, false
		}
		return dispatch{name: args[i+1], rest: args[i+2:], complete: true}, true
	}
	if isCore(cmd, word) {
		return dispatch{}, false
	}
	return dispatch{name: word, rest: args[i+1:]}, true
}

// isCore reports whether word names a core command, an alias, or a Cobra
// built-in. Cobra adds help and completion only inside Execute, so they
// are listed here.
func isCore(cmd *cobra.Command, word string) bool {
	switch word {
	case "help", "completion":
		return true
	}
	for _, c := range cmd.Commands() {
		if c.Name() == word || c.HasAlias(word) {
			return true
		}
	}
	return false
}

// applyHostFlags sets host flags that appear among plugin args. The args
// stay unchanged so the plugin also receives them. Words the host does
// not own are skipped.
func applyHostFlags(pf *pflag.FlagSet, args []string) error {
	for i := 0; i < len(args); {
		if args[i] == "--" {
			return nil
		}
		fl, value, n, known := hostFlag(pf, args, i)
		if !known {
			i++
			continue
		}
		if err := setFlag(fl, value); err != nil {
			return err
		}
		i += n
	}
	return nil
}

// hostFlag parses the flag at args[i] against pf. n is the number of args
// consumed. known is false for non-flags, flags pf does not define, and
// value flags with no value left.
func hostFlag(pf *pflag.FlagSet, args []string, i int) (fl *pflag.Flag, value string, n int, known bool) {
	a := args[i]
	if !strings.HasPrefix(a, "-") || a == "-" || a == "--" {
		return nil, "", 0, false
	}
	var inline *string
	if strings.HasPrefix(a, "--") {
		name := a[2:]
		if eq := strings.IndexByte(name, '='); eq >= 0 {
			v := name[eq+1:]
			inline = &v
			name = name[:eq]
		}
		fl = pf.Lookup(name)
	} else {
		fl = pf.ShorthandLookup(a[1:2])
		if len(a) > 2 {
			v := strings.TrimPrefix(a[2:], "=")
			inline = &v
		}
	}
	if fl == nil {
		return nil, "", 0, false
	}
	switch {
	case inline != nil:
		return fl, *inline, 1, true
	case fl.NoOptDefVal != "":
		return fl, fl.NoOptDefVal, 1, true
	case i+1 < len(args):
		return fl, args[i+1], 2, true
	}
	return nil, "", 0, false
}

func setFlag(fl *pflag.Flag, value string) error {
	if err := fl.Value.Set(value); err != nil {
		return err
	}
	fl.Changed = true
	return nil
}
