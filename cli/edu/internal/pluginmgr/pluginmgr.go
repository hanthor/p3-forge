// Package pluginmgr manages installed edu plugins.
//
// A plugin lives in $XDG_DATA_HOME/edu/plugins/<name>/ as the binary
// edu-<name> and the manifest edu-plugin.json. The host reads manifests
// for listing and help. It executes a plugin only to run it, to complete
// its arguments, or to read its manifest at install time.
package pluginmgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"

	sdkplugin "github.com/LF-Certification/p3-forge/cli/edu/sdk/plugin"
)

// ManifestFile is the manifest name inside a plugin directory.
const ManifestFile = "edu-plugin.json"

// Installed is one plugin found in the plugin directory.
type Installed struct {
	Manifest sdkplugin.Manifest
	// Dir is the plugin directory: <plugins>/<name>.
	Dir string
	// Binary is the executable path: <plugins>/<name>/edu-<name>.
	Binary string
}

// NotFoundError reports a plugin that is not installed. The host exits 2.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("unknown command %q; run \"edu plugin list\"", e.Name)
}

// IncompatibleError reports a plugin whose requires_host range excludes the
// running host. The host exits 2.
type IncompatibleError struct {
	Name     string
	Requires string
	Host     string
}

func (e *IncompatibleError) Error() string {
	return fmt.Sprintf("plugin %q requires host %s; this edu is %s", e.Name, e.Requires, e.Host)
}

// Dir returns the plugin directory. EDU_PLUGIN_DIR overrides it for tests.
func Dir() (string, error) {
	if d := os.Getenv("EDU_PLUGIN_DIR"); d != "" {
		return d, nil
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "edu", "plugins"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve plugin dir: %w", err)
	}
	return filepath.Join(home, ".local", "share", "edu", "plugins"), nil
}

// BinaryName returns the executable name of a plugin.
func BinaryName(name string) string {
	return "edu-" + name
}

// List returns installed plugins sorted by name. It reads manifests
// only. Directories without a manifest are skipped.
func List() ([]Installed, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read plugin dir: %w", err)
	}
	var out []Installed
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := load(dir, e.Name())
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Manifest.Name < out[j].Manifest.Name })
	return out, nil
}

// Lookup returns the plugin called name. A missing binary yields
// *NotFoundError. A binary without a manifest yields an error that asks
// for a reinstall.
func Lookup(name string) (Installed, error) {
	dir, err := Dir()
	if err != nil {
		return Installed{}, err
	}
	bin := filepath.Join(dir, name, BinaryName(name))
	if _, err := os.Stat(bin); err != nil {
		return Installed{}, &NotFoundError{Name: name}
	}
	p, err := load(dir, name)
	if errors.Is(err, os.ErrNotExist) {
		return Installed{}, fmt.Errorf("plugin %q has no %s; run \"edu plugin install\" again", name, ManifestFile)
	}
	return p, err
}

// load reads <dir>/<name>/edu-plugin.json. Missing files return an error
// wrapping os.ErrNotExist.
func load(dir, name string) (Installed, error) {
	pdir := filepath.Join(dir, name)
	data, err := os.ReadFile(filepath.Join(pdir, ManifestFile))
	if err != nil {
		return Installed{}, err
	}
	var m sdkplugin.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Installed{}, fmt.Errorf("parse %s: %w", filepath.Join(pdir, ManifestFile), err)
	}
	return Installed{Manifest: m, Dir: pdir, Binary: filepath.Join(pdir, BinaryName(name))}, nil
}

// CheckHostVersion verifies that host satisfies the requires_host range.
// A host pre-release tag such as 0.1.0-dev is ignored so development
// builds pass release ranges. Failure yields *IncompatibleError; the name
// is filled by the caller when it is known.
func CheckHostVersion(requires, host string) error {
	c, err := semver.NewConstraint(requires)
	if err != nil {
		return fmt.Errorf("invalid requires_host %q: %w", requires, err)
	}
	v, err := semver.NewVersion(host)
	if err != nil {
		return fmt.Errorf("host version %q is not semver: %w", host, err)
	}
	release, err := v.SetPrerelease("")
	if err != nil {
		return err
	}
	if !c.Check(&release) {
		return &IncompatibleError{Requires: requires, Host: host}
	}
	return nil
}
