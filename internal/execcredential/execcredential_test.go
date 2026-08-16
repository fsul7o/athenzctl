package execcredential

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsul7o/athenzctl/internal/config"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if marker := os.Getenv("HELPER_MARKER"); marker != "" {
		if err := os.WriteFile(marker, []byte("called"), 0o600); err != nil {
			os.Exit(2)
		}
	}
	if certPath := os.Getenv("HELPER_CERT_PATH"); certPath != "" {
		if err := os.WriteFile(certPath, []byte(os.Getenv("HELPER_CERT")), 0o600); err != nil {
			os.Exit(3)
		}
	}
	if keyPath := os.Getenv("HELPER_KEY_PATH"); keyPath != "" {
		if err := os.WriteFile(keyPath, []byte(os.Getenv("HELPER_KEY")), 0o600); err != nil {
			os.Exit(4)
		}
	}
	os.Exit(0)
}

func TestFetchReusesUsableCredentialWithoutExec(t *testing.T) {
	dir := t.TempDir()
	certPEM, keyPEM := testCredential(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	certPath, keyPath := writeCredential(t, dir, certPEM, keyPEM)
	marker := filepath.Join(dir, "called")

	gotCert, gotKey, err := Fetch(testExecConfig(certPath, keyPath, marker, certPEM, keyPEM))
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if string(gotCert) != string(certPEM) || string(gotKey) != string(keyPEM) {
		t.Fatal("Fetch() did not return the existing credential")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("expected exec command not to be called")
	}
}

func TestFetchExecsWhenCredentialCannotBeReused(t *testing.T) {
	tests := []struct {
		name     string
		existing func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string)
	}{
		{
			name: "expired",
			existing: func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
				return writeCredential(t, dir, certPEM, keyPEM)
			},
		},
		{
			name: "not yet valid",
			existing: func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
				return writeCredential(t, dir, certPEM, keyPEM)
			},
		},
		{
			name: "malformed certificate",
			existing: func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
				return writeCredential(t, dir, []byte("not a certificate"), keyPEM)
			},
		},
		{
			name: "mismatched key",
			existing: func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
				_, otherKeyPEM := testCredential(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
				return writeCredential(t, dir, certPEM, otherKeyPEM)
			},
		},
		{
			name: "missing files",
			existing: func(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
				return filepath.Join(dir, "missing-cert.pem"), filepath.Join(dir, "missing-key.pem")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var oldCert, oldKey []byte
			switch tt.name {
			case "expired":
				oldCert, oldKey = testCredential(t, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
			case "not yet valid":
				oldCert, oldKey = testCredential(t, time.Now().Add(time.Hour), time.Now().Add(2*time.Hour))
			default:
				oldCert, oldKey = testCredential(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
			}
			certPath, keyPath := tt.existing(t, dir, oldCert, oldKey)
			newCert, newKey := testCredential(t, time.Now().Add(-time.Hour), time.Now().Add(2*time.Hour))
			marker := filepath.Join(dir, "called")

			gotCert, gotKey, err := Fetch(testExecConfig(certPath, keyPath, marker, newCert, newKey))
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}
			if string(gotCert) != string(newCert) || string(gotKey) != string(newKey) {
				t.Fatal("Fetch() did not return the refreshed credential")
			}
			if _, err := os.Stat(marker); err != nil {
				t.Fatalf("expected exec command to be called: %v", err)
			}
		})
	}
}

func TestFetchReturnsExecError(t *testing.T) {
	dir := t.TempDir()
	cfg := testExecConfig(filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), "", nil, nil)
	cfg.Command = filepath.Join(dir, "does-not-exist")
	if _, _, err := Fetch(cfg); err == nil {
		t.Fatal("Fetch() expected an exec error")
	}
}

func testExecConfig(certPath, keyPath, marker string, certPEM, keyPEM []byte) *config.ExecConfig {
	return &config.ExecConfig{
		Command:  os.Args[0],
		Args:     []string{"-test.run=TestHelperProcess", "--"},
		Env:      map[string]string{"GO_WANT_HELPER_PROCESS": "1", "HELPER_MARKER": marker, "HELPER_CERT": string(certPEM), "HELPER_KEY": string(keyPEM), "HELPER_CERT_PATH": certPath, "HELPER_KEY_PATH": keyPath},
		CertPath: certPath,
		KeyPath:  keyPath,
	}
}

func writeCredential(t *testing.T, dir string, certPEM, keyPEM []byte) (string, string) {
	t.Helper()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return certPath, keyPath
}

func testCredential(t *testing.T, notBefore, notAfter time.Time) ([]byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial := time.Now().UnixNano()
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "test"}, NotBefore: notBefore, NotAfter: notAfter, KeyUsage: x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return certPEM, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
}
