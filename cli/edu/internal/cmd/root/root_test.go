package root

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LF-Certification/p3-forge/cli/edu/internal/cmdutil"
	"github.com/LF-Certification/p3-forge/cli/edu/internal/pluginmgr"
	sdkplugin "github.com/LF-Certification/p3-forge/cli/edu/sdk/plugin"
)

func newFactory(out *bytes.Buffer) *cmdutil.Factory {
	return &cmdutil.Factory{
		IOStreams: cmdutil.IOStreams{In: strings.NewReader(""), Out: out, ErrOut: out},
		Version:   "0.1.0",
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		ok       bool
		plugin   string
		rest     []string
		complete bool
		output   string
		context  string
	}{
		{name: "no args", args: nil},
		{name: "core command", args: []string{"version"}},
		{name: "core after host flags", args: []string{"-o", "json", "context", "list"}},
		{name: "help", args: []string{"help"}},
		{name: "unknown host flag goes to cobra", args: []string{"--nope", "placement"}},
		{name: "help flag goes to cobra", args: []string{"-h"}},
		{name: "complete core", args: []string{"__complete", "context", ""}},
		{name: "complete plugin name goes to cobra", args: []string{"__complete", "pla"}},
		{
			name: "plugin with flags before and after",
			args: []string{"-o", "json", "placement", "targets", "list", "--context=dev", "--limit", "5"},
			ok:   true, plugin: "placement",
			rest:   []string{"targets", "list", "--context=dev", "--limit", "5"},
			output: "json", context: "dev",
		},
		{
			name: "plugin short inline flag",
			args: []string{"-ojson", "placement", "-c", "prod"},
			ok:   true, plugin: "placement", rest: []string{"-c", "prod"},
			output: "json", context: "prod",
		},
		{
			name: "complete plugin",
			args: []string{"__complete", "placement", "targets", ""},
			ok:   true, plugin: "placement", rest: []string{"targets", ""}, complete: true,
			output: "table",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			f := newFactory(&out)
			cmd := NewCmd(f)
			d, ok := split(cmd, tc.args)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (dispatch %+v)", ok, tc.ok, d)
			}
			if !ok {
				return
			}
			if err := applyHostFlags(cmd.PersistentFlags(), d.rest); err != nil {
				t.Fatal(err)
			}
			if d.name != tc.plugin || d.complete != tc.complete || strings.Join(d.rest, "\x00") != strings.Join(tc.rest, "\x00") {
				t.Errorf("dispatch = %+v, want name %q rest %q complete %v", d, tc.plugin, tc.rest, tc.complete)
			}
			if f.OutputFlag != tc.output {
				t.Errorf("OutputFlag = %q, want %q", f.OutputFlag, tc.output)
			}
			if f.ContextFlag != tc.context {
				t.Errorf("ContextFlag = %q, want %q", f.ContextFlag, tc.context)
			}
		})
	}
}

func TestRunUnknownCommandExits2(t *testing.T) {
	t.Setenv("EDU_PLUGIN_DIR", t.TempDir())
	var out bytes.Buffer
	err := Run(context.Background(), newFactory(&out), []string{"nosuch", "--flag"})
	if err == nil || err.Error() != `unknown command "nosuch"; run "edu plugin list"` {
		t.Fatalf("err = %v", err)
	}
	if pluginmgr.ExitCode(err) != 2 {
		t.Errorf("exit = %d, want 2", pluginmgr.ExitCode(err))
	}
}

func TestHelpListsPlugins(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("EDU_PLUGIN_DIR", dir)
	for _, p := range []struct{ name, summary string }{
		{"placement", "Manage placement reference data"},
		{"secret", "Hidden plugin"},
	} {
		pdir := filepath.Join(dir, p.name)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		m := sdkplugin.Manifest{SchemaVersion: "1", Name: p.name, Version: "1.0.0", Summary: p.summary, RequiresHost: ">=0.1.0 <1", Hidden: p.name == "secret"}
		data, _ := json.Marshal(m)
		if err := os.WriteFile(filepath.Join(pdir, pluginmgr.ManifestFile), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	if err := Run(context.Background(), newFactory(&out), []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	cmds := strings.Index(help, "Available Commands:")
	plugins := strings.Index(help, "Plugins:\n  placement")
	flags := strings.Index(help, "Flags:")
	if cmds < 0 || plugins < 0 || flags < 0 || !(cmds < plugins && plugins < flags) {
		t.Fatalf("help sections out of order or missing:\n%s", help)
	}
	if !strings.Contains(help, "Manage placement reference data") {
		t.Errorf("summary missing:\n%s", help)
	}
	if strings.Contains(help, "secret") {
		t.Errorf("hidden plugin listed:\n%s", help)
	}

	out.Reset()
	if err := Run(context.Background(), newFactory(&out), []string{"context", "--help"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Plugins:") {
		t.Errorf("subcommand help lists plugins:\n%s", out.String())
	}
}
