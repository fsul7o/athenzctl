// Package authcache provides a small on-disk cache used by the "ntoken" and
// "copperargos" auth-mode credential-fetching packages: the X.509
// certificate and private key minted for the current athenzctl invocation
// are cached so that, as long as the certificate has not neared its expiry,
// subsequent invocations reuse it instead of contacting ZTS again.
package authcache

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	certFileName = "cert.pem"
	keyFileName  = "key.pem"

	// RenewMargin is how much validity must remain on a cached certificate
	// for it to still be considered usable without renewal.
	RenewMargin = time.Hour
)

// Load reads a cached cert/key pair from dir. ok is false if no usable
// cache exists (missing files, or a certificate that fails to parse).
func Load(dir string) (certPEM, keyPEM []byte, ok bool) {
	certPEM, err := os.ReadFile(filepath.Join(dir, certFileName))
	if err != nil {
		return nil, nil, false
	}
	keyPEM, err = os.ReadFile(filepath.Join(dir, keyFileName))
	if err != nil {
		return nil, nil, false
	}
	if _, err := ParseCert(certPEM); err != nil {
		return nil, nil, false
	}
	return certPEM, keyPEM, true
}

// Valid reports whether certPEM still has at least RenewMargin of validity
// remaining.
func Valid(certPEM []byte) bool {
	cert, err := ParseCert(certPEM)
	if err != nil {
		return false
	}
	return time.Now().Add(RenewMargin).Before(cert.NotAfter)
}

// ParseCert decodes the first PEM block in certPEM as an X.509 certificate.
func ParseCert(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, errors.New("no PEM block found")
	}
	return x509.ParseCertificate(block.Bytes)
}

// Save writes certPEM/keyPEM to dir, creating it (0700) if needed. Files are
// written 0600 since keyPEM is sensitive key material.
func Save(dir string, certPEM, keyPEM []byte) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, certFileName), certPEM, 0o600); err != nil {
		return fmt.Errorf("write cached cert: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, keyFileName), keyPEM, 0o600); err != nil {
		return fmt.Errorf("write cached key: %w", err)
	}
	return nil
}
