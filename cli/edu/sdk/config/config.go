// Package config resolves the EDU_* environment the edu host exports to a
// plugin process.
package config

import (
	"errors"
	"os"
	"strings"

	"github.com/LF-Certification/p3-forge/cli/edu/sdk/output"
)

// Config is the plugin-side view of the host contract (plan section 5).
type Config struct {
	// Context is dev, staging or prod.
	Context string
	// APIBaseURL is the gateway base without a trailing slash.
	APIBaseURL string
	// Output is the renderer the host selected with -o.
	Output output.Mode
	// HostVersion is the semver of the launching edu binary; empty when
	// the plugin runs standalone.
	HostVersion string
	// Reexec is the absolute path of the launching edu binary; empty when
	// the plugin runs standalone.
	Reexec string
}

// Getenv is the lookup FromEnv uses. Tests replace it.
var Getenv = os.Getenv

// FromEnv reads EDU_CONTEXT, EDU_API_BASE_URL, EDU_OUTPUT, EDU_HOST_VERSION
// and EDU_REEXEC. Context and base URL are required: without them the
// plugin cannot reach an API, so the error names the fix.
func FromEnv() (Config, error) {
	cfg := Config{
		Context:     Getenv("EDU_CONTEXT"),
		APIBaseURL:  strings.TrimRight(Getenv("EDU_API_BASE_URL"), "/"),
		HostVersion: Getenv("EDU_HOST_VERSION"),
		Reexec:      Getenv("EDU_REEXEC"),
	}
	var missing []string
	if cfg.Context == "" {
		missing = append(missing, "EDU_CONTEXT")
	}
	if cfg.APIBaseURL == "" {
		missing = append(missing, "EDU_API_BASE_URL")
	}
	if len(missing) > 0 {
		return Config{}, errors.New(strings.Join(missing, " and ") +
			" not set; run this plugin through \"edu <plugin>\" or export the EDU_* variables")
	}
	mode, err := output.ParseMode(Getenv("EDU_OUTPUT"))
	if err != nil {
		return Config{}, err
	}
	cfg.Output = mode
	return cfg, nil
}
