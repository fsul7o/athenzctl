// Package authmode holds small helpers shared by the "ntoken" and
// "copperargos" auth-mode credential-fetching packages
// (internal/ntokenauth, internal/copperargosauth): building a ZTS client
// with no client TLS certificate, resolving the on-disk cache directory,
// and resolving CSR subject/DNS-domain/SPIFFE settings from the context's
// issue-defaults.servicecert (or build-time ldflags overrides).
package authmode

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AthenZ/athenz/clients/go/zts"

	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/tlsutil"
)

// These variables are intentionally strings so release builds can override
// them with go build -ldflags -X, mirroring internal/cli/issue's
// RoleCertDefault* mechanism. An empty value means "use the source default".
// They seed the CSR defaults for the "ntoken"/"copperargos" auth-modes
// (context issue-defaults.servicecert still takes priority when set).
var (
	ServiceCertDefaultDNSDomain                 string
	ServiceCertDefaultSubjectCountry            string
	ServiceCertDefaultSubjectProvince           string
	ServiceCertDefaultSubjectOrganization       string
	ServiceCertDefaultSubjectOrganizationalUnit string
	ServiceCertDefaultSpiffe                    string
	ServiceCertDefaultSpiffeTrustDomain         string
	ServiceCertDefaultConcatIntermediateCert    string
)

// AnonymousZTSClient builds a ZTS client with no client TLS certificate at
// all — only CA verification and proxy settings from ctx apply. Callers add
// their own authentication (e.g. an NToken header) or rely on the request
// body itself (e.g. attestation-data) to prove the caller's identity.
func AnonymousZTSClient(ctx *config.Context) (*zts.ZTSClient, error) {
	if ctx.ZTSURL == "" {
		return nil, errors.New("context is missing zts-url")
	}
	tlsCfg, err := tlsutil.BaseTLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.ZTSServerName != "" {
		tlsCfg.ServerName = ctx.ZTSServerName
	}
	transport, err := tlsutil.Transport(ctx, tlsCfg)
	if err != nil {
		return nil, err
	}
	c := zts.NewClient(ctx.ZTSURL, transport)
	return &c, nil
}

// ResolveCacheDir returns configured, if non-empty, otherwise a default
// under ~/.athenzctl/cache/<context-name>/<mode>.
func ResolveCacheDir(ctx *config.Context, mode, configured string) (string, error) {
	if configured != "" {
		return configured, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, config.DefaultDirName, "cache", ctx.Name, mode), nil
}

// ServiceCertDefaults resolves the CSR subject/DNS-domain/SPIFFE settings
// and the concat-intermediate-cert flag, in priority order: context
// issue-defaults.servicecert (highest) > build-time ldflags overrides
// (ServiceCertDefault*) > hardcoded source defaults (spiffe on,
// concat-intermediate-cert off; subj-c/p/o/ou default to empty and are
// simply omitted from the CSR subject when left unconfigured at every
// layer — see csr.NewSubject). There is no CLI-flag layer here, unlike
// `issue rolecert`, since these auth-modes have no per-command flags of
// their own.
func ServiceCertDefaults(ctx *config.Context) (dnsDomain, subjC, subjP, subjO, subjOU string, spiffe bool, spiffeTrustDomain string, concatIntermediateCert bool, err error) {
	dnsDomain = ServiceCertDefaultDNSDomain
	spiffe = true
	spiffeTrustDomain = ServiceCertDefaultSpiffeTrustDomain
	subjC = ServiceCertDefaultSubjectCountry
	subjP = ServiceCertDefaultSubjectProvince
	subjO = ServiceCertDefaultSubjectOrganization
	subjOU = ServiceCertDefaultSubjectOrganizationalUnit
	if ServiceCertDefaultSpiffe != "" {
		v, parseErr := strconv.ParseBool(ServiceCertDefaultSpiffe)
		if parseErr != nil {
			return "", "", "", "", "", false, "", false, fmt.Errorf("invalid built-in spiffe default %q: %w", ServiceCertDefaultSpiffe, parseErr)
		}
		spiffe = v
	}
	if ServiceCertDefaultConcatIntermediateCert != "" {
		v, parseErr := strconv.ParseBool(ServiceCertDefaultConcatIntermediateCert)
		if parseErr != nil {
			return "", "", "", "", "", false, "", false, fmt.Errorf("invalid built-in concat-intermediate-cert default %q: %w", ServiceCertDefaultConcatIntermediateCert, parseErr)
		}
		concatIntermediateCert = v
	}

	if ctx.IssueDefaults == nil || ctx.IssueDefaults.ServiceCert == nil {
		return dnsDomain, subjC, subjP, subjO, subjOU, spiffe, spiffeTrustDomain, concatIntermediateCert, nil
	}
	d := ctx.IssueDefaults.ServiceCert
	if d.DNSDomain != "" {
		dnsDomain = d.DNSDomain
	}
	if d.SubjectCountry != "" {
		subjC = d.SubjectCountry
	}
	if d.SubjectProvince != "" {
		subjP = d.SubjectProvince
	}
	if d.SubjectOrganization != "" {
		subjO = d.SubjectOrganization
	}
	if d.SubjectOrganizationalUnit != "" {
		subjOU = d.SubjectOrganizationalUnit
	}
	if d.Spiffe != nil {
		spiffe = *d.Spiffe
	}
	if d.SpiffeTrustDomain != "" {
		spiffeTrustDomain = d.SpiffeTrustDomain
	}
	if d.ConcatIntermediateCert != nil {
		concatIntermediateCert = *d.ConcatIntermediateCert
	}
	return dnsDomain, subjC, subjP, subjO, subjOU, spiffe, spiffeTrustDomain, concatIntermediateCert, nil
}
