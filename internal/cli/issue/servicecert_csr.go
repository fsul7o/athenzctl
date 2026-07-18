package issue

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/AthenZ/athenz/libs/go/athenzutils"
)

type csrSigner struct {
	key       crypto.Signer
	algorithm x509.SignatureAlgorithm
}

func newCSRSigner(privateKeyPEM []byte) (*csrSigner, error) {
	key, algorithm, err := athenzutils.ExtractSignerInfo(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &csrSigner{key: key, algorithm: algorithm}, nil
}

func generateRSAPrivateKey(path string) ([]byte, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	}
	keyBytes := pem.EncodeToMemory(block)
	if err := os.WriteFile(path, keyBytes, 0o400); err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func generateCSR(signer *csrSigner, subj pkix.Name, host, instanceURI, ip, spiffeURI string) (string, error) {
	tmpl := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: signer.algorithm,
	}
	if host != "" {
		tmpl.DNSNames = []string{host}
	}
	if spiffeURI != "" {
		if u, err := url.Parse(spiffeURI); err == nil {
			tmpl.URIs = append(tmpl.URIs, u)
		}
	}
	if instanceURI != "" {
		if u, err := url.Parse(instanceURI); err == nil {
			tmpl.URIs = append(tmpl.URIs, u)
		}
	}
	if ip != "" {
		tmpl.IPAddresses = []net.IP{net.ParseIP(ip)}
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, &tmpl, signer.key)
	if err != nil {
		return "", fmt.Errorf("create CSR: %w", err)
	}
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}); err != nil {
		return "", fmt.Errorf("encode CSR: %w", err)
	}
	return buf.String(), nil
}
