package ntokenauth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AthenZ/athenz/clients/go/zts"

	"github.com/fsul7o/athenzctl/internal/authcache"
	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/ntokenauth"
)

func generateKeyAndCert(t *testing.T, notAfter time.Time) (keyPEM, certPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return keyPEM, certPEM
}

func TestFetchReturnsValidCacheWithoutNetworkCall(t *testing.T) {
	dir := t.TempDir()
	keyPEM, certPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))
	if err := authcache.Save(dir, certPEM, keyPEM); err != nil {
		t.Fatal(err)
	}

	ctx := &config.Context{
		Name:     "test",
		CacheDir: dir,
		NTokenAuth: &config.NTokenAuthConfig{
			Domain: "example", Service: "svc", PrivateKeyPath: "unused.pem", KeyID: "0",
		},
	}
	gotCert, gotKey, err := ntokenauth.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotCert) != string(certPEM) || string(gotKey) != string(keyPEM) {
		t.Fatal("expected cached cert/key to be returned unchanged")
	}
}

func TestFetchMintsCertViaNToken(t *testing.T) {
	dir := t.TempDir()
	keyPEM, _ := generateKeyAndCert(t, time.Now().Add(time.Hour))
	keyPath := filepath.Join(dir, "svc.key.pem")
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	_, wantCertPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/instance/example/svc/refresh" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Athenz-Principal-Auth") == "" {
			t.Error("expected an NToken auth header on the refresh request")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(zts.Identity{Name: "example.svc", Certificate: string(wantCertPEM)})
	}))
	defer server.Close()

	cacheDir := filepath.Join(dir, "cache")
	ctx := &config.Context{
		Name:                  "test",
		ZTSURL:                server.URL,
		InsecureSkipTLSVerify: true,
		CacheDir:              cacheDir,
		NTokenAuth: &config.NTokenAuthConfig{
			Domain: "example", Service: "svc", PrivateKeyPath: keyPath, KeyID: "1",
		},
		IssueDefaults: &config.IssueDefaults{ServiceCert: &config.CertificateDefaults{DNSDomain: "example.com"}},
	}

	gotCert, gotKey, err := ntokenauth.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotCert) != string(wantCertPEM) {
		t.Fatalf("cert mismatch:\ngot  %s\nwant %s", gotCert, wantCertPEM)
	}
	if string(gotKey) != string(keyPEM) {
		t.Fatalf("key mismatch")
	}

	cachedCert, cachedKey, ok := authcache.Load(cacheDir)
	if !ok || string(cachedCert) != string(wantCertPEM) || string(cachedKey) != string(keyPEM) {
		t.Fatalf("expected the minted cert/key to be cached")
	}
}

func TestFetchUsesConfiguredNTokenHeader(t *testing.T) {
	dir := t.TempDir()
	keyPEM, _ := generateKeyAndCert(t, time.Now().Add(time.Hour))
	keyPath := filepath.Join(dir, "svc.key.pem")
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	_, wantCertPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Auth") == "" {
			t.Error("expected the NToken auth header under the configured custom name")
		}
		if r.Header.Get("Athenz-Principal-Auth") != "" {
			t.Error("did not expect the default header name to be used")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(zts.Identity{Name: "example.svc", Certificate: string(wantCertPEM)})
	}))
	defer server.Close()

	ctx := &config.Context{
		Name:                  "test",
		ZTSURL:                server.URL,
		InsecureSkipTLSVerify: true,
		CacheDir:              filepath.Join(dir, "cache"),
		NTokenAuth: &config.NTokenAuthConfig{
			Domain: "example", Service: "svc", PrivateKeyPath: keyPath, KeyID: "1", Header: "X-Custom-Auth",
		},
		IssueDefaults: &config.IssueDefaults{ServiceCert: &config.CertificateDefaults{DNSDomain: "example.com"}},
	}

	if _, _, err := ntokenauth.Fetch(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestFetchUsesBuildTimeNTokenHeaderDefault(t *testing.T) {
	original := ntokenauth.NTokenHeaderDefault
	defer func() { ntokenauth.NTokenHeaderDefault = original }()
	ntokenauth.NTokenHeaderDefault = "X-Build-Time-Auth"

	dir := t.TempDir()
	keyPEM, _ := generateKeyAndCert(t, time.Now().Add(time.Hour))
	keyPath := filepath.Join(dir, "svc.key.pem")
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	_, wantCertPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Build-Time-Auth") == "" {
			t.Error("expected the NToken auth header under the build-time default name")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(zts.Identity{Name: "example.svc", Certificate: string(wantCertPEM)})
	}))
	defer server.Close()

	ctx := &config.Context{
		Name:                  "test",
		ZTSURL:                server.URL,
		InsecureSkipTLSVerify: true,
		CacheDir:              filepath.Join(dir, "cache"),
		NTokenAuth: &config.NTokenAuthConfig{
			Domain: "example", Service: "svc", PrivateKeyPath: keyPath, KeyID: "1",
		},
		IssueDefaults: &config.IssueDefaults{ServiceCert: &config.CertificateDefaults{DNSDomain: "example.com"}},
	}

	if _, _, err := ntokenauth.Fetch(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestFetchRequiresConfig(t *testing.T) {
	if _, _, err := ntokenauth.Fetch(&config.Context{Name: "test"}); err == nil {
		t.Fatal("expected an error when ntoken-auth config is missing")
	}
	if _, _, err := ntokenauth.Fetch(&config.Context{
		Name:       "test",
		NTokenAuth: &config.NTokenAuthConfig{Domain: "example"},
	}); err == nil {
		t.Fatal("expected an error when required ntoken-auth fields are missing")
	}
}
