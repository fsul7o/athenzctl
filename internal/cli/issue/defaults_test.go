package issue

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func TestBuildCertificateDefaultsAreSeparate(t *testing.T) {
	originalService := [7]string{
		ServiceCertDefaultDNSDomain,
		ServiceCertDefaultSubjectCountry,
		ServiceCertDefaultSubjectProvince,
		ServiceCertDefaultSubjectOrganization,
		ServiceCertDefaultSubjectOrganizationalUnit,
		ServiceCertDefaultSpiffe,
		ServiceCertDefaultSpiffeTrustDomain,
	}
	originalRole := [7]string{
		RoleCertDefaultDNSDomain,
		RoleCertDefaultSubjectCountry,
		RoleCertDefaultSubjectProvince,
		RoleCertDefaultSubjectOrganization,
		RoleCertDefaultSubjectOrganizationalUnit,
		RoleCertDefaultSpiffe,
		RoleCertDefaultSpiffeTrustDomain,
	}
	defer func() {
		ServiceCertDefaultDNSDomain = originalService[0]
		ServiceCertDefaultSubjectCountry = originalService[1]
		ServiceCertDefaultSubjectProvince = originalService[2]
		ServiceCertDefaultSubjectOrganization = originalService[3]
		ServiceCertDefaultSubjectOrganizationalUnit = originalService[4]
		ServiceCertDefaultSpiffe = originalService[5]
		ServiceCertDefaultSpiffeTrustDomain = originalService[6]
		RoleCertDefaultDNSDomain = originalRole[0]
		RoleCertDefaultSubjectCountry = originalRole[1]
		RoleCertDefaultSubjectProvince = originalRole[2]
		RoleCertDefaultSubjectOrganization = originalRole[3]
		RoleCertDefaultSubjectOrganizationalUnit = originalRole[4]
		RoleCertDefaultSpiffe = originalRole[5]
		RoleCertDefaultSpiffeTrustDomain = originalRole[6]
	}()

	ServiceCertDefaultDNSDomain = "service.example"
	ServiceCertDefaultSubjectProvince = "Tokyo"
	ServiceCertDefaultSubjectOrganization = "Service Org"
	ServiceCertDefaultSpiffe = "false"
	RoleCertDefaultDNSDomain = "role.example"
	RoleCertDefaultSubjectProvince = "California"
	RoleCertDefaultSubjectOrganization = "Role Org"
	RoleCertDefaultSpiffe = "true"

	service, err := buildCertificateDefaults(serviceCertKind)
	if err != nil {
		t.Fatal(err)
	}
	role, err := buildCertificateDefaults(roleCertKind)
	if err != nil {
		t.Fatal(err)
	}
	if service.dnsDomain != "service.example" || service.subjectProvince != "Tokyo" || service.subjectOrganization != "Service Org" || service.spiffe {
		t.Fatalf("unexpected service defaults: %+v", service)
	}
	if role.dnsDomain != "role.example" || role.subjectProvince != "California" || role.subjectOrganization != "Role Org" || !role.spiffe {
		t.Fatalf("unexpected role defaults: %+v", role)
	}
}

func TestResolveCertificateDefaultsPriority(t *testing.T) {
	spiffe := true
	path := filepath.Join(t.TempDir(), "config.yaml")
	config := cfg.New()
	config.CurrentContext = "prod"
	config.Upsert(cfg.Context{
		Name: "prod",
		IssueDefaults: &cfg.IssueDefaults{
			ServiceCert: &cfg.CertificateDefaults{
				DNSDomain:           "context.example",
				SubjectProvince:     "Tokyo",
				SubjectOrganization: "Context Org",
				Spiffe:              &spiffe,
				SpiffeTrustDomain:   "context.trust",
			},
		},
	})
	if err := cfg.Save(path, config); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("dns-domain", "", "")
	cmd.Flags().String("subj-c", "", "")
	cmd.Flags().String("subj-p", "", "")
	cmd.Flags().String("subj-o", "", "")
	cmd.Flags().String("subj-ou", "", "")
	cmd.Flags().Bool("spiffe", true, "")
	cmd.Flags().String("spiffe-trust-domain", "", "")
	if err := cmd.Flags().Set("subj-o", "CLI Org"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("spiffe", "false"); err != nil {
		t.Fatal(err)
	}

	defaults, err := resolveCertificateDefaults(cmd, &cliopts.Options{ConfigPath: path}, serviceCertKind)
	if err != nil {
		t.Fatal(err)
	}
	if defaults.dnsDomain != "context.example" || defaults.subjectProvince != "Tokyo" || defaults.subjectOrganization != "CLI Org" || defaults.spiffe {
		t.Fatalf("unexpected resolved defaults: %+v", defaults)
	}
	if defaults.spiffeTrustDomain != "context.trust" {
		t.Fatalf("context trust domain was not applied: %+v", defaults)
	}

	roleCmd := &cobra.Command{}
	roleCmd.Flags().String("dns-domain", "", "")
	roleCmd.Flags().String("subj-c", "", "")
	roleCmd.Flags().String("subj-p", "", "")
	roleCmd.Flags().String("subj-o", "", "")
	roleCmd.Flags().String("subj-ou", "", "")
	roleCmd.Flags().Bool("spiffe", true, "")
	roleCmd.Flags().String("spiffe-trust-domain", "", "")
	roleDefaults, err := resolveCertificateDefaults(roleCmd, &cliopts.Options{ConfigPath: path}, roleCertKind)
	if err != nil {
		t.Fatal(err)
	}
	if roleDefaults.dnsDomain != "" || roleDefaults.subjectOrganization != "Oath Inc." {
		t.Fatalf("service defaults leaked into role defaults: %+v", roleDefaults)
	}
}

func TestResolveCertificateDefaultsWithoutContext(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("dns-domain", "", "")
	cmd.Flags().String("subj-c", "", "")
	cmd.Flags().String("subj-p", "", "")
	cmd.Flags().String("subj-o", "", "")
	cmd.Flags().String("subj-ou", "", "")
	cmd.Flags().Bool("spiffe", true, "")
	cmd.Flags().String("spiffe-trust-domain", "", "")

	defaults, err := resolveCertificateDefaults(cmd, &cliopts.Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")}, roleCertKind)
	if err != nil {
		t.Fatal(err)
	}
	if defaults.subjectCountry != "US" || defaults.subjectOrganization != "Oath Inc." || defaults.subjectOrganizationalUnit != "Athenz" || !defaults.spiffe {
		t.Fatalf("unexpected source defaults: %+v", defaults)
	}
}
