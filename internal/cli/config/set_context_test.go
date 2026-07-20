package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

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

	cmd = New(&Options{ConfigPath: &path})
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

func TestSetContextRoleCertBundleDefaultOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cmd := New(&Options{ConfigPath: &path})
	cmd.SetArgs([]string{
		"set-context", "prod",
		"--rolecert-cacert-bundle-name", "athenz",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	context := loaded.Find("prod")
	if context == nil || context.IssueDefaults == nil || context.IssueDefaults.RoleCert == nil {
		t.Fatalf("rolecert defaults were not saved: %+v", context)
	}
	if context.IssueDefaults.RoleCert.CACertBundleName != "athenz" {
		t.Fatalf("CACertBundleName = %q, want %q", context.IssueDefaults.RoleCert.CACertBundleName, "athenz")
	}
	if context.IssueDefaults.ServiceCert != nil {
		t.Fatalf("servicecert defaults were unexpectedly created: %+v", context.IssueDefaults.ServiceCert)
	}
}

func TestSetContextCertificateDefaultsCanClearValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cmd := New(&Options{ConfigPath: &path})
	cmd.SetArgs([]string{
		"set-context", "prod",
		"--servicecert-subj-c", "JP",
		"--servicecert-spiffe",
		"--servicecert-concat-intermediate-cert",
		"--servicecert-expiry-time", "10",
		"--rolecert-subj-c", "US",
		"--rolecert-spiffe",
		"--rolecert-concat-intermediate-cert",
		"--rolecert-expiry-time", "20",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	cmd = New(&Options{ConfigPath: &path})
	cmd.SetArgs([]string{
		"set-context", "prod",
		"--servicecert-subj-c=",
		"--servicecert-spiffe=false",
		"--servicecert-concat-intermediate-cert=false",
		"--servicecert-expiry-time=0",
		"--rolecert-subj-c=",
		"--rolecert-spiffe=false",
		"--rolecert-concat-intermediate-cert=false",
		"--rolecert-expiry-time=0",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	loaded, err := cfg.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	context := loaded.Find("prod")
	if context == nil || context.IssueDefaults == nil || context.IssueDefaults.ServiceCert == nil || context.IssueDefaults.RoleCert == nil {
		t.Fatalf("certificate defaults were not saved: %+v", context)
	}
	for name, defaults := range map[string]*cfg.CertificateDefaults{
		"servicecert": context.IssueDefaults.ServiceCert,
		"rolecert":    context.IssueDefaults.RoleCert,
	} {
		if defaults.SubjectCountry != "" || defaults.ExpiryTimeMinutes != 0 {
			t.Fatalf("%s string/int defaults were not cleared: %+v", name, defaults)
		}
		if defaults.Spiffe == nil || *defaults.Spiffe {
			t.Fatalf("%s spiffe=false was not preserved: %+v", name, defaults)
		}
		if defaults.ConcatIntermediateCert == nil || *defaults.ConcatIntermediateCert {
			t.Fatalf("%s concat-intermediate-cert=false was not preserved: %+v", name, defaults)
		}
	}
}

func TestApplyCertificateDefaultFlags(t *testing.T) {
	tests := []struct {
		name  string
		set   func(*testing.T, *cobra.Command)
		check func(*testing.T, *cfg.CertificateDefaults)
	}{
		{
			name: "updates all changed common fields",
			set: func(t *testing.T, cmd *cobra.Command) {
				for _, name := range []string{
					"servicecert-subj-c", "servicecert-subj-p", "servicecert-subj-o", "servicecert-subj-ou",
					"servicecert-spiffe", "servicecert-spiffe-trust-domain", "servicecert-dns-domain",
					"servicecert-concat-intermediate-cert", "servicecert-expiry-time", "servicecert-ip", "servicecert-signer-key-id",
				} {
					markFlagChanged(t, cmd, name)
				}
			},
			check: func(t *testing.T, defaults *cfg.CertificateDefaults) {
				t.Helper()
				if defaults.SubjectCountry != "JP" || defaults.SubjectProvince != "Tokyo" || defaults.SubjectOrganization != "Example Org" || defaults.SubjectOrganizationalUnit != "Platform" {
					t.Fatalf("subject defaults = %+v", defaults)
				}
				if defaults.Spiffe == nil || *defaults.Spiffe || defaults.SpiffeTrustDomain != "example.test" || defaults.DNSDomain != "svc.example.test" {
					t.Fatalf("SPIFFE defaults = %+v", defaults)
				}
				if defaults.ConcatIntermediateCert == nil || !*defaults.ConcatIntermediateCert || defaults.ExpiryTimeMinutes != 30 || defaults.IP != "192.0.2.1" || defaults.SignerKeyID != "key-1" {
					t.Fatalf("certificate defaults = %+v", defaults)
				}
			},
		},
		{
			name: "preserves unchanged fields",
			set: func(t *testing.T, cmd *cobra.Command) {
				markFlagChanged(t, cmd, "servicecert-dns-domain")
			},
			check: func(t *testing.T, defaults *cfg.CertificateDefaults) {
				t.Helper()
				if defaults.DNSDomain != "svc.example.test" || defaults.SubjectCountry != "US" || defaults.Spiffe == nil || !*defaults.Spiffe || defaults.ExpiryTimeMinutes != 15 {
					t.Fatalf("unchanged defaults were modified: %+v", defaults)
				}
			},
		},
	}

	values := certificateDefaultValues{
		subjectCountry:            "JP",
		subjectProvince:           "Tokyo",
		subjectOrganization:       "Example Org",
		subjectOrganizationalUnit: "Platform",
		spiffe:                    false,
		spiffeTrustDomain:         "example.test",
		dnsDomain:                 "svc.example.test",
		concatIntermediateCert:    true,
		expiryTimeMinutes:         30,
		ip:                        "192.0.2.1",
		signerKeyID:               "key-1",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCertificateDefaultsCommand()
			tt.set(t, cmd)
			defaults := &cfg.CertificateDefaults{SubjectCountry: "US", Spiffe: boolPtr(true), ExpiryTimeMinutes: 15}
			if !certificateDefaultFlagsChanged(cmd, "servicecert") {
				t.Fatal("certificateDefaultFlagsChanged() = false, want true")
			}
			applyCertificateDefaultFlags(cmd, defaults, "servicecert", values)
			tt.check(t, defaults)
		})
	}
}

func newCertificateDefaultsCommand() *cobra.Command {
	cmd := &cobra.Command{}
	flags := cmd.Flags()
	flags.String("servicecert-subj-c", "", "")
	flags.String("servicecert-subj-p", "", "")
	flags.String("servicecert-subj-o", "", "")
	flags.String("servicecert-subj-ou", "", "")
	flags.Bool("servicecert-spiffe", true, "")
	flags.String("servicecert-spiffe-trust-domain", "", "")
	flags.String("servicecert-dns-domain", "", "")
	flags.Bool("servicecert-concat-intermediate-cert", false, "")
	flags.Int("servicecert-expiry-time", 0, "")
	flags.String("servicecert-ip", "", "")
	flags.String("servicecert-signer-key-id", "", "")
	return cmd
}

func markFlagChanged(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("flag %q not found", name)
	}
	flag.Changed = true
}

func boolPtr(value bool) *bool { return &value }
