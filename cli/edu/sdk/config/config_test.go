package config

import (
	"strings"
	"testing"

	"github.com/LF-Certification/p3-forge/cli/edu/sdk/output"
)

func stubEnv(t *testing.T, env map[string]string) {
	t.Helper()
	old := Getenv
	Getenv = func(k string) string { return env[k] }
	t.Cleanup(func() { Getenv = old })
}

func TestFromEnvFull(t *testing.T) {
	stubEnv(t, map[string]string{
		"EDU_CONTEXT":      "staging",
		"EDU_API_BASE_URL": "https://api.example.test/",
		"EDU_OUTPUT":       "json",
		"EDU_HOST_VERSION": "0.1.0",
		"EDU_REEXEC":       "/usr/local/bin/edu",
	})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIBaseURL != "https://api.example.test" {
		t.Errorf("trailing slash kept: %q", cfg.APIBaseURL)
	}
	if cfg.Output != output.JSON || cfg.Context != "staging" || cfg.HostVersion != "0.1.0" || cfg.Reexec != "/usr/local/bin/edu" {
		t.Errorf("unexpected config %+v", cfg)
	}
}

func TestFromEnvDefaultsOutputToTable(t *testing.T) {
	stubEnv(t, map[string]string{"EDU_CONTEXT": "dev", "EDU_API_BASE_URL": "http://x"})
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != output.Table {
		t.Errorf("got %q", cfg.Output)
	}
}

func TestFromEnvMissing(t *testing.T) {
	stubEnv(t, map[string]string{"EDU_CONTEXT": "dev"})
	_, err := FromEnv()
	if err == nil || !strings.Contains(err.Error(), "EDU_API_BASE_URL") || !strings.Contains(err.Error(), "edu <plugin>") {
		t.Fatalf("error should name the variable and the fix: %v", err)
	}
	stubEnv(t, map[string]string{})
	_, err = FromEnv()
	if err == nil || !strings.Contains(err.Error(), "EDU_CONTEXT and EDU_API_BASE_URL") {
		t.Fatalf("got %v", err)
	}
}

func TestFromEnvRejectsBadOutput(t *testing.T) {
	stubEnv(t, map[string]string{"EDU_CONTEXT": "dev", "EDU_API_BASE_URL": "http://x", "EDU_OUTPUT": "yaml"})
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected error")
	}
}
