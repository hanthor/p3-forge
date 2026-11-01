package plugin

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

func writeFake(t *testing.T, dir, name, version, summary string) {
	t.Helper()
	pdir := filepath.Join(dir, name)
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, pluginmgr.BinaryName(name)), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := sdkplugin.Manifest{SchemaVersion: "1", Name: name, Version: version, Summary: summary, RequiresHost: ">=0.1.0 <1"}
	data, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(pdir, pluginmgr.ManifestFile), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, outputFlag string, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	f := &cmdutil.Factory{
		IOStreams:  cmdutil.IOStreams{In: strings.NewReader(""), Out: &out, ErrOut: &out},
		Version:    "0.1.0",
		OutputFlag: outputFlag,
	}
	cmd := NewCmd(f)
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out.String()
}

func TestListEmpty(t *testing.T) {
	t.Setenv("EDU_PLUGIN_DIR", t.TempDir())
	got := run(t, "table", "list")
	if strings.TrimSpace(got) != "NAME  VERSION  SUMMARY" {
		t.Errorf("table = %q", got)
	}
	got = run(t, "json", "list")
	if strings.TrimSpace(got) != "[]" {
		t.Errorf("json = %q", got)
	}
}

func TestListTwo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("EDU_PLUGIN_DIR", dir)
	writeFake(t, dir, "placement", "0.3.0", "Placement reference data")
	writeFake(t, dir, "hello", "0.0.1", "Say hello")

	got := run(t, "table", "list")
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("table = %q", got)
	}
	if !strings.HasPrefix(lines[1], "hello ") || !strings.Contains(lines[1], "0.0.1") || !strings.Contains(lines[1], "Say hello") {
		t.Errorf("row 1 = %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "placement ") {
		t.Errorf("row 2 = %q", lines[2])
	}

	var manifests []sdkplugin.Manifest
	if err := json.Unmarshal([]byte(run(t, "json", "list")), &manifests); err != nil {
		t.Fatal(err)
	}
	if len(manifests) != 2 || manifests[0].Name != "hello" || manifests[1].RequiresHost != ">=0.1.0 <1" {
		t.Errorf("json = %+v", manifests)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("EDU_PLUGIN_DIR", dir)
	writeFake(t, dir, "hello", "0.0.1", "Say hello")

	got := run(t, "table", "remove", "hello")
	if strings.TrimSpace(got) != "removed hello" {
		t.Errorf("output = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "hello")); !os.IsNotExist(err) {
		t.Errorf("plugin dir still exists: %v", err)
	}

	var out bytes.Buffer
	f := &cmdutil.Factory{IOStreams: cmdutil.IOStreams{Out: &out, ErrOut: &out}, Version: "0.1.0"}
	cmd := NewCmd(f)
	cmd.SetArgs([]string{"remove", "nosuch"})
	err := cmd.ExecuteContext(context.Background())
	if pluginmgr.ExitCode(err) != 2 {
		t.Errorf("remove nosuch: err = %v, want exit 2", err)
	}
}
