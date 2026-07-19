// Package client builds authenticated ZMS/ZTS clients from an athenzctl
// context. mTLS is the primary auth; contexts with auth-mode: "exec" obtain
// their client cert/key by execing an external command that places them at
// a configured path, then reading the result back, before building the mTLS
// config.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/execcredential"
	"golang.org/x/net/proxy"
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
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: ctx.InsecureSkipTLSVerify, // configured explicitly by the user
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

// Transport returns an HTTP transport configured for the context's TLS and
// proxy settings. A proxy value without a scheme is treated as a SOCKS5
// address for compatibility with zms-cli's -s flag.
func Transport(ctx *config.Context) (*http.Transport, error) {
	tlsCfg, err := TLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: defaultTimeout,
	}
	if ctx.ProxyURL == "" {
		return transport, nil
	}

	proxyURL, err := parseProxyURL(ctx.ProxyURL)
	if err != nil {
		return nil, err
	}
	switch proxyURL.Scheme {
	case "http", "https":
		transport.Proxy = http.ProxyURL(proxyURL)
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if proxyURL.User != nil {
			password, _ := proxyURL.User.Password()
			auth = &proxy.Auth{User: proxyURL.User.Username(), Password: password}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, &net.Dialer{})
		if err != nil {
			return nil, fmt.Errorf("configure SOCKS5 proxy %q: %w", ctx.ProxyURL, err)
		}
		transport.Proxy = nil
		transport.DialContext = func(_ context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q (want socks5, http, or https)", proxyURL.Scheme)
	}
	return transport, nil
}

func parseProxyURL(raw string) (*url.URL, error) {
	if !strings.Contains(raw, "://") {
		if _, _, err := net.SplitHostPort(raw); err != nil {
			return nil, fmt.Errorf("invalid proxy %q: expected host:port or URL: %w", raw, err)
		}
		return &url.URL{Scheme: "socks5", Host: raw}, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy %q: %w", raw, err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("invalid proxy %q: missing host", raw)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme == "socks5" || parsed.Scheme == "socks5h" {
		if _, _, err := net.SplitHostPort(parsed.Host); err != nil {
			return nil, fmt.Errorf("invalid proxy %q: SOCKS5 requires host:port: %w", raw, err)
		}
	}
	if parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("invalid proxy %q: path, query, and fragment are not supported", raw)
	}
	if parsed.User != nil && parsed.User.Username() == "" {
		return nil, fmt.Errorf("invalid proxy %q: username must not be empty", raw)
	}
	return parsed, nil
}

// HTTPClient returns an *http.Client wired for mTLS against the given context.
func HTTPClient(ctx *config.Context) (*http.Client, error) {
	transport, err := Transport(ctx)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}, nil
}
