// Package execcredential runs an external command configured via
// auth-mode: "exec" to obtain a fresh Athenz user X.509 certificate + key.
// Rather than reading the credential back from the command's stdout, it
// follows the common Athenz-ecosystem pattern (e.g. ctyano/athenz-user-cert)
// of tools that place the cert/key at fixed file paths and exit: athenzctl
// execs the command, then reads the result from the configured paths.
package execcredential

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/fsul7o/athenzctl/internal/config"
)

// Fetch execs cfg.Command with cfg.Args and cfg.Env merged onto the current
// process's environment, then reads the resulting cert/key PEM from
// cfg.CertPath/cfg.KeyPath. The command's stdout/stderr are passed through
// directly so any interactive login prompts remain visible to the user.
func Fetch(cfg *config.ExecConfig) (certPEM, keyPEM []byte, err error) {
	if cfg == nil || cfg.Command == "" {
		return nil, nil, errors.New("exec: command is required")
	}
	if cfg.CertPath == "" || cfg.KeyPath == "" {
		return nil, nil, errors.New("exec: cert-path and key-path are required")
	}

	cmd := exec.Command(cfg.Command, cfg.Args...)
	cmd.Env = os.Environ()
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("exec %s: %w", cfg.Command, err)
	}

	certPEM, err = os.ReadFile(cfg.CertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("exec %s: read cert-path %s: %w", cfg.Command, cfg.CertPath, err)
	}
	keyPEM, err = os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("exec %s: read key-path %s: %w", cfg.Command, cfg.KeyPath, err)
	}
	return certPEM, keyPEM, nil
}
