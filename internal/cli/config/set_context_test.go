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
		"--rolecert-subj-c", "US",
		"--rolecert-subj-o", "Role Org",
		"--rolecert-spiffe",
		"--rolecert-dns-domain", "role.example",
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
	if role.SubjectCountry != "US" || role.SubjectOrganization != "Role Org" || role.DNSDomain != "role.example" {
		t.Fatalf("unexpected rolecert defaults: %+v", role)
	}
	if role.Spiffe == nil || !*role.Spiffe {
		t.Fatalf("rolecert spiffe=true was not saved: %+v", role)
	}
}
