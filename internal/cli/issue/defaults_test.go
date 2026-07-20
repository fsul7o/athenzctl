package issue

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func TestBuildCertificateDefaultsBuildTimeOverrides(t *testing.T) {
	original := [12]string{
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
		RoleCertDefaultDNSDomain = original[0]
		RoleCertDefaultSubjectCountry = original[1]
		RoleCertDefaultSubjectProvince = original[2]
		RoleCertDefaultSubjectOrganization = original[3]
		RoleCertDefaultSubjectOrganizationalUnit = original[4]
		RoleCertDefaultSpiffe = original[5]
		RoleCertDefaultSpiffeTrustDomain = original[6]
		RoleCertDefaultConcatIntermediateCert = original[7]
		RoleCertDefaultCACertBundleName = original[8]
		RoleCertDefaultExpiryTime = original[9]
		RoleCertDefaultIP = original[10]
		RoleCertDefaultSignerKeyID = original[11]
	}()

	RoleCertDefaultDNSDomain = "role.example"
	RoleCertDefaultSubjectProvince = "California"
	RoleCertDefaultSubjectOrganization = "Role Org"
	RoleCertDefaultSpiffe = "true"
	RoleCertDefaultConcatIntermediateCert = "false"
	RoleCertDefaultCACertBundleName = "athenz"
	RoleCertDefaultExpiryTime = "10"
	RoleCertDefaultIP = "10.2.2.2"
	RoleCertDefaultSignerKeyID = "role-key"

	role, err := buildCertificateDefaults()
	if err != nil {
		t.Fatal(err)
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
			RoleCert: &cfg.CertificateDefaults{
				DNSDomain:              "context.example",
				SubjectProvince:        "Tokyo",
				SubjectOrganization:    "Context Org",
				Spiffe:                 &spiffe,
				SpiffeTrustDomain:      "context.trust",
				ConcatIntermediateCert: &concat,
				CACertBundleName:       "context-bundle",
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
	cmd.Flags().String("cacert-bundle-name", "", "")
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

	defaults, err := resolveCertificateDefaults(cmd, &cliopts.Options{ConfigPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if defaults.dnsDomain != "context.example" || defaults.subjectProvince != "Tokyo" || defaults.subjectOrganization != "CLI Org" || defaults.spiffe {
		t.Fatalf("unexpected resolved defaults: %+v", defaults)
	}
	if defaults.spiffeTrustDomain != "context.trust" {
		t.Fatalf("context trust domain was not applied: %+v", defaults)
	}
	if !defaults.concatIntermediateCert || defaults.caCertBundleName != "context-bundle" || defaults.expiryTimeMinutes != 15 {
		t.Fatalf("context concat-intermediate-cert/cacert-bundle-name/expiry-time were not applied: %+v", defaults)
	}
	if defaults.ip != "10.9.9.9" {
		t.Fatalf("CLI --ip did not override context default: %+v", defaults)
	}
	if defaults.signerKeyID != "context-key" {
		t.Fatalf("context signer-key-id was not applied: %+v", defaults)
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

	defaults, err := resolveCertificateDefaults(cmd, &cliopts.Options{ConfigPath: filepath.Join(t.TempDir(), "missing.yaml")})
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
