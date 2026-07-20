package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	cfg "github.com/fsul7o/athenzctl/internal/config"
)

const (
	serviceCertKind certificateKind = iota
	roleCertKind
)

// These variables are intentionally strings so release builds can override
// them with go build -ldflags -X. An empty value means "use the source default".
var (
	ServiceCertDefaultDNSDomain                 string
	ServiceCertDefaultSubjectCountry            string
	ServiceCertDefaultSubjectProvince           string
	ServiceCertDefaultSubjectOrganization       string
	ServiceCertDefaultSubjectOrganizationalUnit string
	ServiceCertDefaultSpiffe                    string
	ServiceCertDefaultSpiffeTrustDomain         string
	ServiceCertDefaultConcatIntermediateCert    string
	ServiceCertDefaultExpiryTime                string
	ServiceCertDefaultIP                        string
	ServiceCertDefaultSignerKeyID               string
	RoleCertDefaultDNSDomain                    string
	RoleCertDefaultSubjectCountry               string
	RoleCertDefaultSubjectProvince              string
	RoleCertDefaultSubjectOrganization          string
	RoleCertDefaultSubjectOrganizationalUnit    string
	RoleCertDefaultSpiffe                       string
	RoleCertDefaultSpiffeTrustDomain            string
	RoleCertDefaultConcatIntermediateCert       string
	RoleCertDefaultCACertBundleName             string
	RoleCertDefaultExpiryTime                   string
	RoleCertDefaultIP                           string
	RoleCertDefaultSignerKeyID                  string
)

type certificateKind uint8

type certificateDefaults struct {
	dnsDomain                 string
	subjectCountry            string
	subjectProvince           string
	subjectOrganization       string
	subjectOrganizationalUnit string
	spiffe                    bool
	spiffeTrustDomain         string
	concatIntermediateCert    bool
	caCertBundleName          string
	expiryTimeMinutes         int
	ip                        string
	signerKeyID               string
}

func resolveCertificateDefaults(cmd *cobra.Command, opts *cliopts.Options, kind certificateKind) (certificateDefaults, error) {
	defaults, err := buildCertificateDefaults(kind)
	if err != nil {
		return certificateDefaults{}, err
	}

	ctx, err := opts.LoadSelectedContext()
	if err != nil {
		return certificateDefaults{}, err
	}
	if ctx != nil && ctx.IssueDefaults != nil {
		var configured *cfg.CertificateDefaults
		if kind == serviceCertKind {
			if ctx.IssueDefaults != nil {
				configured = ctx.IssueDefaults.ServiceCert
			}
		} else {
			if ctx.IssueDefaults != nil {
				configured = ctx.IssueDefaults.RoleCert
			}
		}
		applyConfiguredCertificateDefaults(&defaults, configured)
	}

	if cmd.Flags().Changed("dns-domain") {
		defaults.dnsDomain, err = cmd.Flags().GetString("dns-domain")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("subj-c") {
		defaults.subjectCountry, err = cmd.Flags().GetString("subj-c")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("subj-p") {
		defaults.subjectProvince, err = cmd.Flags().GetString("subj-p")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("subj-o") {
		defaults.subjectOrganization, err = cmd.Flags().GetString("subj-o")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("subj-ou") {
		defaults.subjectOrganizationalUnit, err = cmd.Flags().GetString("subj-ou")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("spiffe") {
		defaults.spiffe, err = cmd.Flags().GetBool("spiffe")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("spiffe-trust-domain") {
		defaults.spiffeTrustDomain, err = cmd.Flags().GetString("spiffe-trust-domain")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("concat-intermediate-cert") {
		defaults.concatIntermediateCert, err = cmd.Flags().GetBool("concat-intermediate-cert")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("cacert-bundle-name") {
		defaults.caCertBundleName, err = cmd.Flags().GetString("cacert-bundle-name")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("expiry-time") {
		defaults.expiryTimeMinutes, err = cmd.Flags().GetInt("expiry-time")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("ip") {
		defaults.ip, err = cmd.Flags().GetString("ip")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	if cmd.Flags().Changed("signer-key-id") {
		defaults.signerKeyID, err = cmd.Flags().GetString("signer-key-id")
		if err != nil {
			return certificateDefaults{}, err
		}
	}
	return defaults, nil
}

func buildCertificateDefaults(kind certificateKind) (certificateDefaults, error) {
	defaults := certificateDefaults{
		dnsDomain:                 "",
		subjectCountry:            "US",
		subjectProvince:           "",
		subjectOrganization:       "Oath Inc.",
		subjectOrganizationalUnit: "Athenz",
		spiffe:                    true,
		spiffeTrustDomain:         "",
		concatIntermediateCert:    false,
		caCertBundleName:          "",
		expiryTimeMinutes:         0,
		ip:                        "",
		signerKeyID:               "",
	}
	var spiffeOverride, concatOverride, expiryOverride string
	if kind == serviceCertKind {
		applyBuildOverride(&defaults.dnsDomain, ServiceCertDefaultDNSDomain)
		applyBuildOverride(&defaults.subjectCountry, ServiceCertDefaultSubjectCountry)
		applyBuildOverride(&defaults.subjectProvince, ServiceCertDefaultSubjectProvince)
		applyBuildOverride(&defaults.subjectOrganization, ServiceCertDefaultSubjectOrganization)
		applyBuildOverride(&defaults.subjectOrganizationalUnit, ServiceCertDefaultSubjectOrganizationalUnit)
		spiffeOverride = ServiceCertDefaultSpiffe
		applyBuildOverride(&defaults.spiffeTrustDomain, ServiceCertDefaultSpiffeTrustDomain)
		concatOverride = ServiceCertDefaultConcatIntermediateCert
		expiryOverride = ServiceCertDefaultExpiryTime
		applyBuildOverride(&defaults.ip, ServiceCertDefaultIP)
		applyBuildOverride(&defaults.signerKeyID, ServiceCertDefaultSignerKeyID)
	} else {
		applyBuildOverride(&defaults.dnsDomain, RoleCertDefaultDNSDomain)
		applyBuildOverride(&defaults.subjectCountry, RoleCertDefaultSubjectCountry)
		applyBuildOverride(&defaults.subjectProvince, RoleCertDefaultSubjectProvince)
		applyBuildOverride(&defaults.subjectOrganization, RoleCertDefaultSubjectOrganization)
		applyBuildOverride(&defaults.subjectOrganizationalUnit, RoleCertDefaultSubjectOrganizationalUnit)
		spiffeOverride = RoleCertDefaultSpiffe
		applyBuildOverride(&defaults.spiffeTrustDomain, RoleCertDefaultSpiffeTrustDomain)
		concatOverride = RoleCertDefaultConcatIntermediateCert
		applyBuildOverride(&defaults.caCertBundleName, RoleCertDefaultCACertBundleName)
		expiryOverride = RoleCertDefaultExpiryTime
		applyBuildOverride(&defaults.ip, RoleCertDefaultIP)
		applyBuildOverride(&defaults.signerKeyID, RoleCertDefaultSignerKeyID)
	}
	if spiffeOverride != "" {
		spiffe, err := strconv.ParseBool(spiffeOverride)
		if err != nil {
			return certificateDefaults{}, fmt.Errorf("invalid built-in spiffe default %q: %w", spiffeOverride, err)
		}
		defaults.spiffe = spiffe
	}
	if concatOverride != "" {
		concat, err := strconv.ParseBool(concatOverride)
		if err != nil {
			return certificateDefaults{}, fmt.Errorf("invalid built-in concat-intermediate-cert default %q: %w", concatOverride, err)
		}
		defaults.concatIntermediateCert = concat
	}
	if expiryOverride != "" {
		expiry, err := strconv.Atoi(expiryOverride)
		if err != nil {
			return certificateDefaults{}, fmt.Errorf("invalid built-in expiry-time default %q: %w", expiryOverride, err)
		}
		defaults.expiryTimeMinutes = expiry
	}
	return defaults, nil
}

func applyConfiguredCertificateDefaults(defaults *certificateDefaults, configured *cfg.CertificateDefaults) {
	if configured == nil {
		return
	}
	applyBuildOverride(&defaults.dnsDomain, configured.DNSDomain)
	applyBuildOverride(&defaults.subjectCountry, configured.SubjectCountry)
	applyBuildOverride(&defaults.subjectProvince, configured.SubjectProvince)
	applyBuildOverride(&defaults.subjectOrganization, configured.SubjectOrganization)
	applyBuildOverride(&defaults.subjectOrganizationalUnit, configured.SubjectOrganizationalUnit)
	applyBuildOverride(&defaults.spiffeTrustDomain, configured.SpiffeTrustDomain)
	if configured.Spiffe != nil {
		defaults.spiffe = *configured.Spiffe
	}
	if configured.ConcatIntermediateCert != nil {
		defaults.concatIntermediateCert = *configured.ConcatIntermediateCert
	}
	applyBuildOverride(&defaults.caCertBundleName, configured.CACertBundleName)
	if configured.ExpiryTimeMinutes != 0 {
		defaults.expiryTimeMinutes = configured.ExpiryTimeMinutes
	}
	applyBuildOverride(&defaults.ip, configured.IP)
	applyBuildOverride(&defaults.signerKeyID, configured.SignerKeyID)
}

func applyBuildOverride(destination *string, override string) {
	if override != "" {
		*destination = override
	}
}
