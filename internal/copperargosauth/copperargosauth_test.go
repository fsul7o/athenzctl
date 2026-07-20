package copperargosauth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
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
	"github.com/fsul7o/athenzctl/internal/copperargosauth"
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
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
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

func baseContext(serverURL, cacheDir string) *config.Context {
	return &config.Context{
		Name:                  "test",
		ZTSURL:                serverURL,
		InsecureSkipTLSVerify: true,
		CacheDir:              cacheDir,
		CopperArgosAuth: &config.CopperArgosAuthConfig{
			Domain: "example", Service: "svc", Provider: "sys.auth.zts", Instance: "i-1",
		},
		IssueDefaults: &config.IssueDefaults{ServiceCert: &config.CertificateDefaults{DNSDomain: "example.com"}},
	}
}

func TestFetchReturnsValidCacheWithoutNetworkCall(t *testing.T) {
	dir := t.TempDir()
	keyPEM, certPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))
	if err := authcache.Save(dir, certPEM, keyPEM); err != nil {
		t.Fatal(err)
	}

	ctx := baseContext("", dir)
	ctx.CopperArgosAuth.AttestationDataPath = "unused.data"

	gotCert, gotKey, err := copperargosauth.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotCert) != string(certPEM) || string(gotKey) != string(keyPEM) {
		t.Fatal("expected cached cert/key to be returned unchanged")
	}
}

func TestFetchSelfRefreshesExpiredCache(t *testing.T) {
	dir := t.TempDir()
	keyPEM, oldCertPEM := generateKeyAndCert(t, time.Now().Add(-time.Minute)) // already expired
	cacheDir := filepath.Join(dir, "cache")
	if err := authcache.Save(cacheDir, oldCertPEM, keyPEM); err != nil {
		t.Fatal(err)
	}
	_, newCertPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/instance/example/svc/refresh" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(r.TLS.PeerCertificates) == 0 {
			t.Error("expected the cached cert to be presented as the client certificate")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(zts.Identity{Name: "example.svc", Certificate: string(newCertPEM)})
	}))
	server.TLS = &tls.Config{ClientAuth: tls.RequestClientCert}
	server.StartTLS()
	defer server.Close()

	ctx := baseContext(server.URL, cacheDir)
	ctx.CopperArgosAuth.AttestationDataPath = "unused.data"

	gotCert, gotKey, err := copperargosauth.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotCert) != string(newCertPEM) {
		t.Fatalf("cert mismatch:\ngot  %s\nwant %s", gotCert, newCertPEM)
	}
	if string(gotKey) != string(keyPEM) {
		t.Fatal("expected the CSR key to be reused across self-refresh")
	}
}

func TestFetchRegistersWhenNoCacheExists(t *testing.T) {
	dir := t.TempDir()
	attestationPath := filepath.Join(dir, "attestation.data")
	if err := os.WriteFile(attestationPath, []byte("fake-attestation-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, wantCertPEM := generateKeyAndCert(t, time.Now().Add(24*time.Hour))

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/instance" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(r.TLS.PeerCertificates) != 0 {
			t.Error("expected an anonymous (no client cert) connection for registration")
		}
		var req zts.InstanceRegisterInformation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.AttestationData != "fake-attestation-data" {
			t.Fatalf("unexpected attestation data: %q", req.AttestationData)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(zts.InstanceIdentity{
			Provider: "sys.auth.zts", Name: "example.svc", InstanceId: "i-1",
			X509Certificate: string(wantCertPEM),
		})
	}))
	defer server.Close()

	cacheDir := filepath.Join(dir, "cache")
	ctx := baseContext(server.URL, cacheDir)
	ctx.CopperArgosAuth.AttestationDataPath = attestationPath

	gotCert, _, err := copperargosauth.Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotCert) != string(wantCertPEM) {
		t.Fatalf("cert mismatch:\ngot  %s\nwant %s", gotCert, wantCertPEM)
	}
	if _, _, ok := authcache.Load(cacheDir); !ok {
		t.Fatal("expected the registered cert/key to be cached")
	}
}

func TestFetchRequiresAttestationDataFile(t *testing.T) {
	dir := t.TempDir()
	ctx := baseContext("", filepath.Join(dir, "cache"))
	ctx.CopperArgosAuth.AttestationDataPath = filepath.Join(dir, "missing.data")

	if _, _, err := copperargosauth.Fetch(ctx); err == nil {
		t.Fatal("expected an error when attestation-data-path does not exist")
	}
}

func TestFetchRequiresConfig(t *testing.T) {
	if _, _, err := copperargosauth.Fetch(&config.Context{Name: "test"}); err == nil {
		t.Fatal("expected an error when copperargos-auth config is missing")
	}
}
