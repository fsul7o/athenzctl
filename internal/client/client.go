// Package client builds authenticated ZMS/ZTS clients from an athenzctl
// context. mTLS is the primary auth; contexts with auth-mode: "exec" obtain
// their client cert/key by execing an external command that places them at
// a configured path, then reading the result back, before building the mTLS
// config.
package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/execcredential"
)

const defaultTimeout = 30 * time.Second

// TLSConfig builds a *tls.Config for the context. For auth-mode "exec" the
// client cert/key are obtained in-memory from the configured exec plugin;
// otherwise they're loaded from the context's cert/key file paths.
func TLSConfig(ctx *config.Context) (*tls.Config, error) {
	var cert tls.Certificate
	switch ctx.AuthMode {
	case "exec":
		if ctx.Exec == nil {
			return nil, errors.New("auth-mode is \"exec\" but context has no exec config")
		}
		certPEM, keyPEM, err := execcredential.Fetch(ctx.Exec)
		if err != nil {
			return nil, fmt.Errorf("exec credential: %w", err)
		}
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse exec credential cert/key: %w", err)
		}
	default:
		if ctx.Cert == "" || ctx.Key == "" {
			return nil, errors.New("context is missing cert/key (set auth-mode to \"exec\" to obtain credentials from an external command)")
		}
		var err error
		cert, err = tls.LoadX509KeyPair(ctx.Cert, ctx.Key)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key: %w", err)
		}
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	if ctx.CACert != "" {
		pem, err := os.ReadFile(ctx.CACert)
		if err != nil {
			return nil, fmt.Errorf("read ca-cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("ca-cert %q contained no valid PEM", ctx.CACert)
		}
		tlsCfg.RootCAs = pool
	}
	return tlsCfg, nil
}

// HTTPClient returns an *http.Client wired for mTLS against the given context.
func HTTPClient(ctx *config.Context) (*http.Client, error) {
	tlsCfg, err := TLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout: defaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}, nil
}
