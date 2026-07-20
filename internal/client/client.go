// Package client builds authenticated ZMS/ZTS clients from an athenzctl
// context. mTLS is the primary auth; contexts with auth-mode: "exec" obtain
// their client cert/key by execing an external command that places them at
// a configured path, then reading the result back, before building the mTLS
// config.
package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/copperargosauth"
	"github.com/fsul7o/athenzctl/internal/execcredential"
	"github.com/fsul7o/athenzctl/internal/ntokenauth"
	"github.com/fsul7o/athenzctl/internal/tlsutil"
)

// TLSConfig builds a *tls.Config for the context. For auth-mode "exec" the
// client cert/key are obtained in-memory from the configured exec plugin;
// for "ntoken"/"copperargos" they are minted on the fly (and cached) via
// ZTS; otherwise they're loaded from the context's cert/key file paths.
func TLSConfig(ctx *config.Context) (*tls.Config, error) {
	tlsCfg, err := tlsutil.BaseTLSConfig(ctx)
	if err != nil {
		return nil, err
	}

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
	case "ntoken":
		certPEM, keyPEM, err := ntokenauth.Fetch(ctx)
		if err != nil {
			return nil, fmt.Errorf("ntoken-auth credential: %w", err)
		}
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse ntoken-auth cert/key: %w", err)
		}
	case "copperargos":
		certPEM, keyPEM, err := copperargosauth.Fetch(ctx)
		if err != nil {
			return nil, fmt.Errorf("copperargos-auth credential: %w", err)
		}
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse copperargos-auth cert/key: %w", err)
		}
	default:
		if ctx.Cert == "" || ctx.Key == "" {
			return nil, errors.New("context is missing cert/key (set auth-mode to \"exec\" to obtain credentials from an external command)")
		}
		cert, err = tls.LoadX509KeyPair(ctx.Cert, ctx.Key)
		if err != nil {
			return nil, fmt.Errorf("load client cert/key: %w", err)
		}
	}
	tlsCfg.Certificates = []tls.Certificate{cert}
	return tlsCfg, nil
}

// Transport returns an HTTP transport configured for the context's TLS and
// proxy settings. A proxy value without a scheme is treated as a SOCKS5
// address for compatibility with zms-cli's -s flag.
func Transport(ctx *config.Context) (*http.Transport, error) {
	tlsCfg, err := TLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return tlsutil.Transport(ctx, tlsCfg)
}

// HTTPClient returns an *http.Client wired for mTLS against the given context.
func HTTPClient(ctx *config.Context) (*http.Client, error) {
	transport, err := Transport(ctx)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout:   tlsutil.DefaultTimeout,
		Transport: transport,
	}, nil
}
