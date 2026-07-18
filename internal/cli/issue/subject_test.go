package issue

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestNewCSRSubjectOmitsEmptyAttributes(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer := &csrSigner{key: key, algorithm: x509.SHA256WithRSA}
	subj := newCSRSubject("service.example", "", "", "", "")
	tests := []struct {
		name string
		make func() (string, error)
	}{
		{
			name: "service certificate",
			make: func() (string, error) {
				return generateCSR(signer, subj, "service.example", "", "", "")
			},
		},
		{
			name: "role certificate",
			make: func() (string, error) {
				return generateRoleCSR(signer, subj, "service.example", "example.service", "example", "", "")
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
			csr, err := x509.ParseCertificateRequest(block.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			if got := csr.Subject.String(); got != "CN=service.example" {
				t.Fatalf("unexpected subject: %q", got)
			}
		})
	}
}

func TestNewCSRSubjectIncludesNonEmptyAttributes(t *testing.T) {
	subj := newCSRSubject("service.example", "US", "Tokyo", "Example Org", "Athenz")

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
