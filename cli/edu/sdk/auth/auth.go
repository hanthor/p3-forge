// Package auth obtains the access token a plugin sends to the API.
package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Token returns EDU_TOKEN when the host exported it. Otherwise it runs
// `<EDU_REEXEC> auth token`, falling back to `edu` on PATH, so a plugin
// works standalone during development. The host exports an empty
// EDU_TOKEN when the user is logged out, so this is the normal path for
// the login hint. The child's stderr is captured and returned as the
// error, with the child's `edu: ` prefix removed, so the plugin prints
// the message once.
func Token(ctx context.Context) (string, error) {
	if tok := os.Getenv("EDU_TOKEN"); tok != "" {
		return tok, nil
	}
	exe := os.Getenv("EDU_REEXEC")
	if exe == "" {
		exe = "edu"
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, exe, "auth", "token")
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			msg := strings.TrimPrefix(strings.TrimSpace(stderr.String()), "edu: ")
			if msg == "" {
				msg = fmt.Sprintf("%s auth token: %s", exe, exitErr)
			}
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("run %s auth token: %w", exe, err)
	}
	tok := strings.TrimSpace(stdout.String())
	if tok == "" {
		return "", fmt.Errorf("%s auth token printed no token; run \"edu auth login\"", exe)
	}
	return tok, nil
}
