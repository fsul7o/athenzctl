package csr_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/fsul7o/athenzctl/internal/csr"
)

func TestNewSubjectOmitsEmptyAttributes(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer := &csr.Signer{Key: key, Algorithm: x509.SHA256WithRSA}
	subj := csr.NewSubject("service.example", "", "", "", "")
	tests := []struct {
		name string
		make func() (string, error)
	}{
		{
			name: "service certificate",
			make: func() (string, error) {
				return csr.GenerateServiceCSR(signer, subj, "service.example", "", "", "")
			},
		},
		{
			name: "role certificate",
			make: func() (string, error) {
				return csr.GenerateRoleCSR(signer, subj, "service.example", "example.service", "example", "", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csrPEM, err := tt.make()
			if err != nil {
				t.Fatal(err)
			}
			block, _ := pem.Decode([]byte(csrPEM))
			if block == nil {
				t.Fatal("generated CSR did not contain a PEM block")
			}
			req, err := x509.ParseCertificateRequest(block.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			if got := req.Subject.String(); got != "CN=service.example" {
				t.Fatalf("unexpected subject: %q", got)
			}
		})
	}
}

func TestNewSubjectIncludesNonEmptyAttributes(t *testing.T) {
	subj := csr.NewSubject("service.example", "US", "Tokyo", "Example Org", "Athenz")

	if len(subj.Country) != 1 || subj.Country[0] != "US" {
		t.Fatalf("unexpected country: %#v", subj.Country)
	}
	if len(subj.Province) != 1 || subj.Province[0] != "Tokyo" {
		t.Fatalf("unexpected province: %#v", subj.Province)
	}
	if len(subj.Organization) != 1 || subj.Organization[0] != "Example Org" {
		t.Fatalf("unexpected organization: %#v", subj.Organization)
	}
	if len(subj.OrganizationalUnit) != 1 || subj.OrganizationalUnit[0] != "Athenz" {
		t.Fatalf("unexpected organizational unit: %#v", subj.OrganizationalUnit)
	}
}
