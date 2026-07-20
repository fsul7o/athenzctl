// Package csr generates X.509 certificate signing requests for Athenz
// service and role certificates. It intentionally has no dependency on
// cobra, cliopts, or the client package, so it can be imported both by CLI
// commands (internal/cli/issue) and by lower-level credential-fetching
// packages (e.g. auth-mode implementations under internal/client) without
// creating import cycles.
package csr

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
	"strings"

	"github.com/AthenZ/athenz/libs/go/athenzutils"
)

// Signer wraps a private key and its signature algorithm for CSR generation.
type Signer struct {
	Key       crypto.Signer
	Algorithm x509.SignatureAlgorithm
}

// NewSigner extracts a crypto.Signer and its signature algorithm from
// PEM-encoded private key bytes.
func NewSigner(privateKeyPEM []byte) (*Signer, error) {
	key, algorithm, err := athenzutils.ExtractSignerInfo(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &Signer{Key: key, Algorithm: algorithm}, nil
}

// NewSubject builds a pkix.Name for a CSR subject. Per RFC 5280, an
// attribute is only included when non-empty; CommonName is always set.
func NewSubject(commonName, country, province, organization, organizationalUnit string) pkix.Name {
	subj := pkix.Name{CommonName: commonName}
	if country != "" {
		subj.Country = []string{country}
	}
	if province != "" {
		subj.Province = []string{province}
	}
	if organization != "" {
		subj.Organization = []string{organization}
	}
	if organizationalUnit != "" {
		subj.OrganizationalUnit = []string{organizationalUnit}
	}
	return subj
}

// GenerateRSAPrivateKeyPEM generates a new 2048-bit RSA private key
// in-memory and returns its PEM-encoded (PKCS#1) bytes without writing it
// anywhere.
func GenerateRSAPrivateKeyPEM() ([]byte, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	}
	return pem.EncodeToMemory(block), nil
}

// ServiceIdentityNames computes the conventional SAN DNS host, CSR subject
// CommonName, and (optional) SPIFFE URI for an Athenz service X.509
// certificate, given the service's domain, name, and DNS suffix. This
// mirrors the naming convention used by zts-svccert.
func ServiceIdentityNames(domain, service, dnsDomain, spiffeTrustDomain string, spiffe bool) (host, commonName, spiffeURI string) {
	hyphenDomain := strings.ReplaceAll(domain, ".", "-")
	host = fmt.Sprintf("%s.%s.%s", service, hyphenDomain, dnsDomain)
	commonName = fmt.Sprintf("%s.%s", domain, service)
	if spiffe || spiffeTrustDomain != "" {
		if spiffeTrustDomain != "" {
			spiffeURI = fmt.Sprintf("spiffe://%s/ns/default/sa/%s", spiffeTrustDomain, commonName)
		} else {
			spiffeURI = fmt.Sprintf("spiffe://%s/sa/%s", domain, service)
		}
	}
	return host, commonName, spiffeURI
}

// GenerateServiceCSR builds a service identity CSR (mirrors zts-svccert).
func GenerateServiceCSR(signer *Signer, subj pkix.Name, host, instanceURI, ip, spiffeURI string) (string, error) {
	tmpl := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: signer.Algorithm,
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
	return encodeCSR(&tmpl, signer.Key)
}

// GenerateRoleCSR builds a role certificate CSR (mirrors zts-rolecert).
func GenerateRoleCSR(signer *Signer, subj pkix.Name, host, principal, dnsDomain, ip, spiffeURI string) (string, error) {
	tmpl := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: signer.Algorithm,
	}
	if host != "" {
		tmpl.DNSNames = []string{host}
	}
	if ip != "" {
		tmpl.IPAddresses = []net.IP{net.ParseIP(ip)}
	}
	tmpl.EmailAddresses = []string{fmt.Sprintf("%s@%s", principal, dnsDomain)}
	if spiffeURI != "" {
		if u, err := url.Parse(spiffeURI); err == nil {
			tmpl.URIs = append(tmpl.URIs, u)
		}
	}
	if u, err := url.Parse(fmt.Sprintf("athenz://principal/%s", principal)); err == nil {
		tmpl.URIs = append(tmpl.URIs, u)
	}
	return encodeCSR(&tmpl, signer.Key)
}

func encodeCSR(tmpl *x509.CertificateRequest, key crypto.Signer) (string, error) {
	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, key)
	if err != nil {
		return "", fmt.Errorf("create CSR: %w", err)
	}
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}); err != nil {
		return "", fmt.Errorf("encode CSR: %w", err)
	}
	return buf.String(), nil
}
