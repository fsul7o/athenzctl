package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsul7o/athenzctl/internal/config"
)

func testContext(t *testing.T) *config.Context {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          new(big.Int).SetInt64(1),
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "client.cert.pem")
	keyPath := filepath.Join(dir, "client.key.pem")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600); err != nil {
		t.Fatal(err)
	}
	return &config.Context{Cert: certPath, Key: keyPath}
}

func TestTLSConfigHonorsInsecureSkipTLSVerify(t *testing.T) {
	ctx := testContext(t)
	ctx.InsecureSkipTLSVerify = true
	tlsConfig, err := TLSConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !tlsConfig.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify was not enabled")
	}
}

func TestTransportProxyForms(t *testing.T) {
	tests := []struct {
		name       string
		proxy      string
		wantScheme string
		wantHost   string
		wantSOCKS  bool
	}{
		{name: "bare socks5", proxy: "127.0.0.1:1080", wantScheme: "socks5", wantHost: "127.0.0.1:1080", wantSOCKS: true},
		{name: "socks5 url", proxy: "socks5://user:pass@proxy.example:1080", wantScheme: "socks5", wantHost: "proxy.example:1080", wantSOCKS: true},
		{name: "http", proxy: "http://user:pass@proxy.example:8080", wantScheme: "http", wantHost: "proxy.example:8080"},
		{name: "https", proxy: "https://proxy.example:8443", wantScheme: "https", wantHost: "proxy.example:8443"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testContext(t)
			ctx.ProxyURL = tt.proxy
			transport, err := Transport(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantSOCKS {
				if transport.DialContext == nil || transport.Proxy != nil {
					t.Fatalf("unexpected SOCKS transport: proxyConfigured=%v dialContext=%v", transport.Proxy != nil, transport.DialContext != nil)
				}
				return
			}
			if transport.DialContext != nil {
				t.Fatal("HTTP transport unexpectedly has a SOCKS dialer")
			}
			proxyURL, err := transport.Proxy(&http.Request{URL: mustURL(t, "https://zms.example")})
			if err != nil {
				t.Fatal(err)
			}
			if proxyURL == nil || proxyURL.Scheme != tt.wantScheme || proxyURL.Host != tt.wantHost {
				t.Fatalf("proxy URL = %v, want %s://%s", proxyURL, tt.wantScheme, tt.wantHost)
			}
			if tt.name == "http" {
				if user := proxyURL.User.Username(); user != "user" {
					t.Fatalf("proxy username = %q, want user", user)
				}
			}
		})
	}
}

func TestTransportRejectsInvalidProxy(t *testing.T) {
	for _, proxyURL := range []string{"proxy.example", "ftp://proxy.example:21", "http://", "socks5://proxy.example", "http://proxy.example/path"} {
		t.Run(proxyURL, func(t *testing.T) {
			ctx := testContext(t)
			ctx.ProxyURL = proxyURL
			if _, err := Transport(ctx); err == nil {
				t.Fatalf("Transport(%q) succeeded, want error", proxyURL)
			}
		})
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
