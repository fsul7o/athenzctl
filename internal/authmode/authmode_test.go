package authmode_test

import (
	"testing"

	"github.com/fsul7o/athenzctl/internal/authmode"
	"github.com/fsul7o/athenzctl/internal/config"
)

func TestServiceCertDefaultsBuildTimeOverride(t *testing.T) {
	original := [7]string{
		authmode.ServiceCertDefaultDNSDomain,
		authmode.ServiceCertDefaultSubjectCountry,
		authmode.ServiceCertDefaultSubjectProvince,
		authmode.ServiceCertDefaultSubjectOrganization,
		authmode.ServiceCertDefaultSubjectOrganizationalUnit,
		authmode.ServiceCertDefaultSpiffe,
		authmode.ServiceCertDefaultSpiffeTrustDomain,
	}
	defer func() {
		authmode.ServiceCertDefaultDNSDomain = original[0]
		authmode.ServiceCertDefaultSubjectCountry = original[1]
		authmode.ServiceCertDefaultSubjectProvince = original[2]
		authmode.ServiceCertDefaultSubjectOrganization = original[3]
		authmode.ServiceCertDefaultSubjectOrganizationalUnit = original[4]
		authmode.ServiceCertDefaultSpiffe = original[5]
		authmode.ServiceCertDefaultSpiffeTrustDomain = original[6]
	}()

	authmode.ServiceCertDefaultDNSDomain = "build.example"
	authmode.ServiceCertDefaultSubjectProvince = "Osaka"
	authmode.ServiceCertDefaultSubjectOrganization = "Build Org"
	authmode.ServiceCertDefaultSpiffe = "false"
	authmode.ServiceCertDefaultSpiffeTrustDomain = "build.trust"

	dnsDomain, subjC, subjP, subjO, subjOU, spiffe, spiffeTrustDomain, _, err := authmode.ServiceCertDefaults(&config.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if dnsDomain != "build.example" || subjP != "Osaka" || subjO != "Build Org" || spiffe {
		t.Fatalf("unexpected build-time defaults: dnsDomain=%q subjP=%q subjO=%q spiffe=%v", dnsDomain, subjP, subjO, spiffe)
	}
	if subjC != "" || subjOU != "" {
		t.Fatalf("unconfigured subj-c/subj-ou should stay empty, got: subjC=%q subjOU=%q", subjC, subjOU)
	}
	if spiffeTrustDomain != "build.trust" {
		t.Fatalf("unexpected build-time spiffe-trust-domain: %q", spiffeTrustDomain)
	}
}

func TestServiceCertDefaultsSubjectFieldsDefaultToEmpty(t *testing.T) {
	// With no build-time ldflags and no context override, subj-c/p/o/ou
	// must come out as empty strings (and thus be omitted from the CSR
	// subject by csr.NewSubject) rather than falling back to a hardcoded
	// non-empty value.
	dnsDomain, subjC, subjP, subjO, subjOU, _, _, _, err := authmode.ServiceCertDefaults(&config.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if subjC != "" || subjP != "" || subjO != "" || subjOU != "" {
		t.Fatalf("expected all subject fields to default to empty, got: subjC=%q subjP=%q subjO=%q subjOU=%q", subjC, subjP, subjO, subjOU)
	}
	if dnsDomain != "" {
		t.Fatalf("expected dns-domain to default to empty, got: %q", dnsDomain)
	}
}

func TestServiceCertDefaultsContextOverridesBuildTime(t *testing.T) {
	original := authmode.ServiceCertDefaultDNSDomain
	defer func() { authmode.ServiceCertDefaultDNSDomain = original }()
	authmode.ServiceCertDefaultDNSDomain = "build.example"

	spiffe := false
	ctx := &config.Context{
		IssueDefaults: &config.IssueDefaults{
			ServiceCert: &config.CertificateDefaults{
				DNSDomain: "context.example",
				Spiffe:    &spiffe,
			},
		},
	}
	dnsDomain, _, _, _, _, gotSpiffe, _, _, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if dnsDomain != "context.example" {
		t.Fatalf("context dns-domain did not override build-time default: %q", dnsDomain)
	}
	if gotSpiffe {
		t.Fatal("context spiffe=false was not applied")
	}
}

func TestServiceCertDefaultsRejectsInvalidBuildTimeSpiffe(t *testing.T) {
	original := authmode.ServiceCertDefaultSpiffe
	defer func() { authmode.ServiceCertDefaultSpiffe = original }()
	authmode.ServiceCertDefaultSpiffe = "not-a-bool"

	if _, _, _, _, _, _, _, _, err := authmode.ServiceCertDefaults(&config.Context{}); err == nil {
		t.Fatal("expected an error for an invalid built-in spiffe default")
	}
}

func TestServiceCertDefaultsConcatIntermediateCertBuildTimeOverride(t *testing.T) {
	original := authmode.ServiceCertDefaultConcatIntermediateCert
	defer func() { authmode.ServiceCertDefaultConcatIntermediateCert = original }()
	authmode.ServiceCertDefaultConcatIntermediateCert = "true"

	_, _, _, _, _, _, _, concatIntermediateCert, err := authmode.ServiceCertDefaults(&config.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if !concatIntermediateCert {
		t.Fatal("expected build-time concat-intermediate-cert default to be applied")
	}
}

func TestServiceCertDefaultsConcatIntermediateCertContextOverridesBuildTime(t *testing.T) {
	original := authmode.ServiceCertDefaultConcatIntermediateCert
	defer func() { authmode.ServiceCertDefaultConcatIntermediateCert = original }()
	authmode.ServiceCertDefaultConcatIntermediateCert = "true"

	concat := false
	ctx := &config.Context{
		IssueDefaults: &config.IssueDefaults{
			ServiceCert: &config.CertificateDefaults{
				ConcatIntermediateCert: &concat,
			},
		},
	}
	_, _, _, _, _, _, _, concatIntermediateCert, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if concatIntermediateCert {
		t.Fatal("context concat-intermediate-cert=false did not override the build-time default")
	}
}

func TestServiceCertDefaultsRejectsInvalidBuildTimeConcatIntermediateCert(t *testing.T) {
	original := authmode.ServiceCertDefaultConcatIntermediateCert
	defer func() { authmode.ServiceCertDefaultConcatIntermediateCert = original }()
	authmode.ServiceCertDefaultConcatIntermediateCert = "not-a-bool"

	if _, _, _, _, _, _, _, _, err := authmode.ServiceCertDefaults(&config.Context{}); err == nil {
		t.Fatal("expected an error for an invalid built-in concat-intermediate-cert default")
	}
}
