package config

import (
	"path/filepath"
	"testing"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func TestSetContextIssueDefaultsAreIndependent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cmd := New(&Options{ConfigPath: &path})
	cmd.SetArgs([]string{
		"set-context", "prod",
		"--servicecert-subj-c", "JP",
		"--servicecert-subj-o", "Service Org",
		"--servicecert-spiffe=false",
		"--servicecert-dns-domain", "service.example",
		"--servicecert-concat-intermediate-cert",
		"--servicecert-expiry-time", "5",
		"--servicecert-ip", "10.1.1.1",
		"--servicecert-signer-key-id", "svc-key",
		"--rolecert-subj-c", "US",
		"--rolecert-subj-o", "Role Org",
		"--rolecert-spiffe",
		"--rolecert-dns-domain", "role.example",
		"--rolecert-concat-intermediate-cert",
		"--rolecert-cacert-bundle-name", "athenz",
		"--rolecert-expiry-time", "10",
		"--rolecert-ip", "10.2.2.2",
		"--rolecert-signer-key-id", "role-key",
		"--insecure-skip-tls-verify",
		"--proxy", "http://proxy.example:8080",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	context := loaded.Find("prod")
	if context == nil || context.IssueDefaults == nil {
		t.Fatalf("issue defaults were not saved: %+v", context)
	}
	service := context.IssueDefaults.ServiceCert
	role := context.IssueDefaults.RoleCert
	if service == nil || role == nil {
		t.Fatalf("certificate defaults were not separated: %+v", context.IssueDefaults)
	}
	if service.SubjectCountry != "JP" || service.SubjectOrganization != "Service Org" || service.DNSDomain != "service.example" {
		t.Fatalf("unexpected servicecert defaults: %+v", service)
	}
	if service.Spiffe == nil || *service.Spiffe {
		t.Fatalf("servicecert spiffe=false was not saved: %+v", service)
	}
	if service.ConcatIntermediateCert == nil || !*service.ConcatIntermediateCert {
		t.Fatalf("servicecert concat-intermediate-cert was not saved: %+v", service)
	}
	if service.ExpiryTimeMinutes != 5 || service.IP != "10.1.1.1" || service.SignerKeyID != "svc-key" {
		t.Fatalf("unexpected servicecert cert-detail defaults: %+v", service)
	}
	if role.SubjectCountry != "US" || role.SubjectOrganization != "Role Org" || role.DNSDomain != "role.example" {
		t.Fatalf("unexpected rolecert defaults: %+v", role)
	}
	if role.Spiffe == nil || !*role.Spiffe {
		t.Fatalf("rolecert spiffe=true was not saved: %+v", role)
	}
	if role.ConcatIntermediateCert == nil || !*role.ConcatIntermediateCert || role.CACertBundleName != "athenz" {
		t.Fatalf("rolecert concat-intermediate-cert/cacert-bundle-name were not saved: %+v", role)
	}
	if role.ExpiryTimeMinutes != 10 || role.IP != "10.2.2.2" || role.SignerKeyID != "role-key" {
		t.Fatalf("unexpected rolecert cert-detail defaults: %+v", role)
	}
	if !context.InsecureSkipTLSVerify || context.ProxyURL != "http://proxy.example:8080" {
		t.Fatalf("connection settings were not saved: %+v", context)
	}

	cmd.SetArgs([]string{
		"set-context", "prod",
		"--proxy", "socks5://proxy.example:1080",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	loaded, err = cfg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	context = loaded.Find("prod")
	if context == nil || !context.InsecureSkipTLSVerify || context.ProxyURL != "socks5://proxy.example:1080" || context.IssueDefaults == nil {
		t.Fatalf("updating connection settings lost existing context data: %+v", context)
	}
}
