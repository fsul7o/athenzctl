// Package execcredential obtains an Athenz user X.509 certificate + key for
// auth-mode: "exec". It reuses a currently usable credential at the
// configured paths and runs the external command only when those files need
// to be refreshed. The command follows the common Athenz-ecosystem pattern
// (e.g. ctyano/athenz-user-cert) of placing cert/key files at fixed paths.
package execcredential

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/fsul7o/athenzctl/internal/config"
)

// Fetch returns a currently usable credential from cfg.CertPath/cfg.KeyPath
// when one exists. If the credential is missing, malformed, expired, not yet
// valid, or does not match its private key, Fetch execs cfg.Command with
// cfg.Args and cfg.Env merged onto the current process's environment, then
// reads the resulting cert/key PEM from the configured paths. The command's
// stdout/stderr are passed through directly so any interactive login prompts
// remain visible to the user.
func Fetch(cfg *config.ExecConfig) (certPEM, keyPEM []byte, err error) {
	if cfg == nil || cfg.Command == "" {
		return nil, nil, errors.New("exec: command is required")
	}
	if cfg.CertPath == "" || cfg.KeyPath == "" {
		return nil, nil, errors.New("exec: cert-path and key-path are required")
	}

	if certPEM, keyPEM, ok := usableCredential(cfg.CertPath, cfg.KeyPath, time.Now()); ok {
		return certPEM, keyPEM, nil
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

// usableCredential validates the existing certificate and private key as a
// pair. Errors are intentionally treated as a cache miss: the configured
// command is responsible for repairing missing or invalid output files.
func usableCredential(certPath, keyPath string, now time.Time) ([]byte, []byte, bool) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, false
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, false
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil || now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return nil, nil, false
	}
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return nil, nil, false
	}
	return certPEM, keyPEM, true
}
