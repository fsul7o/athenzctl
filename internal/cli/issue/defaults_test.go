package issue

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func TestBuildCertificateDefaultsAreSeparate(t *testing.T) {
	originalService := [11]string{
		ServiceCertDefaultDNSDomain,
		ServiceCertDefaultSubjectCountry,
		ServiceCertDefaultSubjectProvince,
		ServiceCertDefaultSubjectOrganization,
		ServiceCertDefaultSubjectOrganizationalUnit,
		ServiceCertDefaultSpiffe,
		ServiceCertDefaultSpiffeTrustDomain,
		ServiceCertDefaultConcatIntermediateCert,
		ServiceCertDefaultExpiryTime,
		ServiceCertDefaultIP,
		ServiceCertDefaultSignerKeyID,
	}
	originalRole := [13]string{
		RoleCertDefaultDNSDomain,
		RoleCertDefaultSubjectCountry,
		RoleCertDefaultSubjectProvince,
		RoleCertDefaultSubjectOrganization,
		RoleCertDefaultSubjectOrganizationalUnit,
		RoleCertDefaultSpiffe,
		RoleCertDefaultSpiffeTrustDomain,
		RoleCertDefaultConcatIntermediateCert,
		RoleCertDefaultCACertBundleName,
		RoleCertDefaultExpiryTime,
		RoleCertDefaultIP,
		RoleCertDefaultSignerKeyID,
	}
	defer func() {
		ServiceCertDefaultDNSDomain = originalService[0]
		ServiceCertDefaultSubjectCountry = originalService[1]
		ServiceCertDefaultSubjectProvince = originalService[2]
		ServiceCertDefaultSubjectOrganization = originalService[3]
		ServiceCertDefaultSubjectOrganizationalUnit = originalService[4]
		ServiceCertDefaultSpiffe = originalService[5]
		ServiceCertDefaultSpiffeTrustDomain = originalService[6]
		ServiceCertDefaultConcatIntermediateCert = originalService[7]
		ServiceCertDefaultExpiryTime = originalService[8]
		ServiceCertDefaultIP = originalService[9]
		ServiceCertDefaultSignerKeyID = originalService[10]
		RoleCertDefaultDNSDomain = originalRole[0]
		RoleCertDefaultSubjectCountry = originalRole[1]
		RoleCertDefaultSubjectProvince = originalRole[2]
		RoleCertDefaultSubjectOrganization = originalRole[3]
		RoleCertDefaultSubjectOrganizationalUnit = originalRole[4]
		RoleCertDefaultSpiffe = originalRole[5]
		RoleCertDefaultSpiffeTrustDomain = originalRole[6]
		RoleCertDefaultConcatIntermediateCert = originalRole[7]
		RoleCertDefaultCACertBundleName = originalRole[8]
		RoleCertDefaultExpiryTime = originalRole[9]
		RoleCertDefaultIP = originalRole[10]
		RoleCertDefaultSignerKeyID = originalRole[11]
	}()

	ServiceCertDefaultDNSDomain = "service.example"
	ServiceCertDefaultSubjectProvince = "Tokyo"
	ServiceCertDefaultSubjectOrganization = "Service Org"
	ServiceCertDefaultSpiffe = "false"
	ServiceCertDefaultConcatIntermediateCert = "true"
	ServiceCertDefaultExpiryTime = "5"
	ServiceCertDefaultIP = "10.1.1.1"
	ServiceCertDefaultSignerKeyID = "svc-key"
	RoleCertDefaultDNSDomain = "role.example"
	RoleCertDefaultSubjectProvince = "California"
	RoleCertDefaultSubjectOrganization = "Role Org"
	RoleCertDefaultSpiffe = "true"
	RoleCertDefaultConcatIntermediateCert = "false"
	RoleCertDefaultCACertBundleName = "athenz"
	RoleCertDefaultExpiryTime = "10"
	RoleCertDefaultIP = "10.2.2.2"
	RoleCertDefaultSignerKeyID = "role-key"

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
	if !service.concatIntermediateCert || service.expiryTimeMinutes != 5 || service.ip != "10.1.1.1" || service.signerKeyID != "svc-key" {
		t.Fatalf("unexpected service cert-detail defaults: %+v", service)
	}
	if role.dnsDomain != "role.example" || role.subjectProvince != "California" || role.subjectOrganization != "Role Org" || !role.spiffe {
		t.Fatalf("unexpected role defaults: %+v", role)
	}
	if role.concatIntermediateCert || role.caCertBundleName != "athenz" || role.expiryTimeMinutes != 10 || role.ip != "10.2.2.2" || role.signerKeyID != "role-key" {
		t.Fatalf("unexpected role cert-detail defaults: %+v", role)
	}
}

func TestResolveCertificateDefaultsPriority(t *testing.T) {
	spiffe := true
	concat := true
	path := filepath.Join(t.TempDir(), "config.yaml")
	config := cfg.New()
	config.CurrentContext = "prod"
	config.Upsert(cfg.Context{
		Name: "prod",
		IssueDefaults: &cfg.IssueDefaults{
			ServiceCert: &cfg.CertificateDefaults{
				DNSDomain:              "context.example",
				SubjectProvince:        "Tokyo",
				SubjectOrganization:    "Context Org",
				Spiffe:                 &spiffe,
				SpiffeTrustDomain:      "context.trust",
				ConcatIntermediateCert: &concat,
				ExpiryTimeMinutes:      15,
				IP:                     "172.16.0.1",
				SignerKeyID:            "context-key",
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
	cmd.Flags().Bool("concat-intermediate-cert", false, "")
	cmd.Flags().Int("expiry-time", 0, "")
	cmd.Flags().String("ip", "", "")
	cmd.Flags().String("signer-key-id", "", "")
	if err := cmd.Flags().Set("subj-o", "CLI Org"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("spiffe", "false"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("ip", "10.9.9.9"); err != nil {
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
	if !defaults.concatIntermediateCert || defaults.expiryTimeMinutes != 15 {
		t.Fatalf("context concat-intermediate-cert/expiry-time were not applied: %+v", defaults)
	}
	if defaults.ip != "10.9.9.9" {
		t.Fatalf("CLI --ip did not override context default: %+v", defaults)
	}
	if defaults.signerKeyID != "context-key" {
		t.Fatalf("context signer-key-id was not applied: %+v", defaults)
	}

	roleCmd := &cobra.Command{}
	roleCmd.Flags().String("dns-domain", "", "")
	roleCmd.Flags().String("subj-c", "", "")
	roleCmd.Flags().String("subj-p", "", "")
	roleCmd.Flags().String("subj-o", "", "")
	roleCmd.Flags().String("subj-ou", "", "")
	roleCmd.Flags().Bool("spiffe", true, "")
	roleCmd.Flags().String("spiffe-trust-domain", "", "")
	roleCmd.Flags().Bool("concat-intermediate-cert", false, "")
	roleCmd.Flags().String("cacert-bundle-name", "", "")
	roleCmd.Flags().Int("expiry-time", 0, "")
	roleCmd.Flags().String("ip", "", "")
	roleCmd.Flags().String("signer-key-id", "", "")
	roleDefaults, err := resolveCertificateDefaults(roleCmd, &cliopts.Options{ConfigPath: path}, roleCertKind)
	if err != nil {
		t.Fatal(err)
	}
	if roleDefaults.dnsDomain != "" || roleDefaults.subjectOrganization != "Oath Inc." {
		t.Fatalf("service defaults leaked into role defaults: %+v", roleDefaults)
	}
	if roleDefaults.concatIntermediateCert || roleDefaults.caCertBundleName != "" || roleDefaults.expiryTimeMinutes != 0 || roleDefaults.ip != "" || roleDefaults.signerKeyID != "" {
		t.Fatalf("service cert-detail defaults leaked into role defaults: %+v", roleDefaults)
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
	if defaults.concatIntermediateCert || defaults.caCertBundleName != "" || defaults.expiryTimeMinutes != 0 || defaults.ip != "" || defaults.signerKeyID != "" {
		t.Fatalf("unexpected source cert-detail defaults: %+v", defaults)
	}
}
