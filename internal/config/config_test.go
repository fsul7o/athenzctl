package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsEmpty(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Contexts) != 0 {
		t.Fatalf("expected empty contexts, got %d", len(cfg.Contexts))
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	orig := New()
	orig.Upsert(Context{Name: "prod", ZMSURL: "https://zms.example", ZTSURL: "https://zts.example", Cert: "/x/c.pem", Key: "/x/k.pem"})
	orig.CurrentContext = "prod"

	if err := Save(path, orig); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CurrentContext != "prod" {
		t.Fatalf("current-context lost: %+v", loaded)
	}
	got, err := loaded.Current()
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if got.ZMSURL != "https://zms.example" {
		t.Fatalf("zms url mismatch: %q", got.ZMSURL)
	}
}

func TestIssueDefaultsRoundTripKeepsCertificateTypesSeparate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	serviceSpiffe := false
	roleSpiffe := true
	orig := New()
	orig.Upsert(Context{
		Name: "prod",
		IssueDefaults: &IssueDefaults{
			ServiceCert: &CertificateDefaults{
				DNSDomain:                 "service.example",
				SubjectCountry:            "JP",
				SubjectProvince:           "Tokyo",
				SubjectOrganization:       "Service Org",
				SubjectOrganizationalUnit: "Service Unit",
				Spiffe:                    &serviceSpiffe,
				SpiffeTrustDomain:         "service.trust",
			},
			RoleCert: &CertificateDefaults{
				DNSDomain:                 "role.example",
				SubjectCountry:            "US",
				SubjectProvince:           "California",
				SubjectOrganization:       "Role Org",
				SubjectOrganizationalUnit: "Role Unit",
				Spiffe:                    &roleSpiffe,
				SpiffeTrustDomain:         "role.trust",
			},
		},
	})
	if err := Save(path, orig); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := loaded.Find("prod")
	if ctx == nil || ctx.IssueDefaults == nil || ctx.IssueDefaults.ServiceCert == nil || ctx.IssueDefaults.RoleCert == nil {
		t.Fatalf("issue defaults were not loaded: %+v", ctx)
	}
	if ctx.IssueDefaults.ServiceCert.DNSDomain != "service.example" || ctx.IssueDefaults.RoleCert.DNSDomain != "role.example" {
		t.Fatalf("certificate defaults were mixed: %+v", ctx.IssueDefaults)
	}
	if ctx.IssueDefaults.ServiceCert.SubjectProvince != "Tokyo" || ctx.IssueDefaults.RoleCert.SubjectProvince != "California" {
		t.Fatalf("certificate provinces were not preserved: %+v", ctx.IssueDefaults)
	}
	if ctx.IssueDefaults.ServiceCert.Spiffe == nil || *ctx.IssueDefaults.ServiceCert.Spiffe {
		t.Fatalf("servicecert spiffe=false was not preserved: %+v", ctx.IssueDefaults.ServiceCert)
	}
	if ctx.IssueDefaults.RoleCert.Spiffe == nil || !*ctx.IssueDefaults.RoleCert.Spiffe {
		t.Fatalf("rolecert spiffe=true was not preserved: %+v", ctx.IssueDefaults.RoleCert)
	}
}

func TestUpsertReplacesByName(t *testing.T) {
	c := New()
	c.Upsert(Context{Name: "a", ZMSURL: "one"})
	c.Upsert(Context{Name: "a", ZMSURL: "two"})
	if len(c.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(c.Contexts))
	}
	if c.Contexts[0].ZMSURL != "two" {
		t.Fatalf("upsert did not replace: %q", c.Contexts[0].ZMSURL)
	}
}

func TestRemoveClearsCurrent(t *testing.T) {
	c := New()
	c.Upsert(Context{Name: "a"})
	c.CurrentContext = "a"
	if !c.Remove("a") {
		t.Fatal("Remove returned false for existing context")
	}
	if c.CurrentContext != "" {
		t.Fatalf("current-context not cleared: %q", c.CurrentContext)
	}
	if c.Remove("missing") {
		t.Fatal("Remove returned true for missing context")
	}
}

func TestCurrentUnsetError(t *testing.T) {
	c := New()
	if _, err := c.Current(); err == nil {
		t.Fatal("expected error when current-context unset")
	}
	c.CurrentContext = "nope"
	if _, err := c.Current(); err == nil {
		t.Fatal("expected error when current-context points to missing entry")
	}
}
