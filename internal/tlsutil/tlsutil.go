// Package tlsutil builds low-level TLS configuration (CA verification and
// proxy handling) shared by internal/client and the credential-fetching
// auth-mode packages (e.g. ntoken/copperargos) that cannot import
// internal/client without creating an import cycle (those packages are in
// turn imported by internal/client to implement new auth-mode cases).
//
// Building a client TLS certificate (if any) is left to the caller: some
// auth flows need no client certificate at all (anonymous TLS, relying on
// an NToken header or on attestation-data instead), others need an ad-hoc
// certificate that never comes from the on-disk athenzctl context.
package tlsutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"github.com/fsul7o/athenzctl/internal/config"
)

// DefaultTimeout is the response header timeout applied to transports built
// by this package.
const DefaultTimeout = 30 * time.Second

// BaseTLSConfig builds a *tls.Config from the context's CA verification
// settings only (CACert, InsecureSkipTLSVerify). It sets no client
// certificate; callers append tlsCfg.Certificates themselves if the auth
// flow requires one.
func BaseTLSConfig(ctx *config.Context) (*tls.Config, error) {
	tlsCfg := &tls.Config{
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

// Transport returns an HTTP transport using tlsCfg and configured for the
// context's proxy settings. A proxy value without a scheme is treated as a
// SOCKS5 address for compatibility with zms-cli's -s flag.
func Transport(ctx *config.Context, tlsCfg *tls.Config) (*http.Transport, error) {
	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: DefaultTimeout,
	}
	if ctx.ProxyURL == "" {
		return transport, nil
	}

	proxyURL, err := ParseProxyURL(ctx.ProxyURL)
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

// ParseProxyURL parses raw as a proxy URL, treating a bare host:port as
// SOCKS5 for compatibility with zms-cli's -s flag.
func ParseProxyURL(raw string) (*url.URL, error) {
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
