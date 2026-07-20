package issue

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zts"
	rdl "github.com/ardielle/ardielle-go/rdl"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// newRoleCert requests a role X.509 certificate. Flags mirror `zts-rolecert`
// (with `--` prefix); ZTS URL and mTLS credentials come from the athenzctl
// context. The caller's service domain/service is auto-extracted from the
// context client cert Subject CN if not overridden with --domain / --service.
func newRoleCert(opts *cliopts.Options) *cobra.Command {
	var (
		roleDomain                  string
		roleName                    string
		service                     string
		roleKeyFile                 string
		dnsDomain                   string
		subjC, subjP, subjO, subjOU string
		spiffe                      bool
		spiffeTrustDomain           string
		ip                          string
		oldRoleCertPath             string
		csrOnly                     bool
		expiryTime                  int
		outPath                     string
		proxyForPrincipal           string
		concatIntermediateCert      bool
		caCertBundleName            string
		signerKeyID                 string
	)
	cmd := &cobra.Command{
		Use:   "rolecert",
		Short: "Request an X.509 role certificate from ZTS",
		Long: `Request an X.509 role certificate for the given role.

Flags mirror zts-rolecert. --domain is the caller's service domain (defaults
to the CN extracted from the context client cert); --role-domain is the
domain that owns the requested role.

The CSR is generated in-process from --role-key-file (defaulting to the
context's service key). Pass --csr to print the CSR and exit without
calling ZTS.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			defaults, err := resolveCertificateDefaults(cmd, opts, roleCertKind)
			if err != nil {
				return err
			}
			dnsDomain = defaults.dnsDomain
			subjC = defaults.subjectCountry
			subjP = defaults.subjectProvince
			subjO = defaults.subjectOrganization
			subjOU = defaults.subjectOrganizationalUnit
			spiffe = defaults.spiffe
			spiffeTrustDomain = defaults.spiffeTrustDomain
			concatIntermediateCert = defaults.concatIntermediateCert
			caCertBundleName = defaults.caCertBundleName
			expiryTime = defaults.expiryTimeMinutes
			ip = defaults.ip
			signerKeyID = defaults.signerKeyID

			if roleDomain == "" {
				return errors.New("issue rolecert requires --role-domain")
			}
			if roleName == "" {
				return errors.New("issue rolecert requires --role-name")
			}
			if dnsDomain == "" {
				return errors.New("issue rolecert requires --dns-domain or a rolecert default")
			}

			ctx, err := opts.LoadContext()
			if err != nil {
				return err
			}
			domain := opts.Domain
			if domain == "" || service == "" {
				certDomain, certService, err := extractServiceFromCertFile(ctx.Cert)
				if err != nil {
					return fmt.Errorf("resolve caller identity from context cert: %w", err)
				}
				if domain == "" {
					domain = certDomain
				}
				if service == "" {
					service = certService
				}
			}
			if domain == "" || service == "" {
				return errors.New("could not determine caller domain/service; pass --domain and --service")
			}

			keyPath := roleKeyFile
			if keyPath == "" {
				keyPath = ctx.Key
			}
			keyBytes, err := os.ReadFile(keyPath)
			if err != nil {
				return fmt.Errorf("read role key %s: %w", keyPath, err)
			}
			signer, err := newCSRSigner(keyBytes)
			if err != nil {
				return fmt.Errorf("load role key: %w", err)
			}

			hyphenDomain := strings.ReplaceAll(domain, ".", "-")
			host := fmt.Sprintf("%s.%s.%s", service, hyphenDomain, dnsDomain)
			principal := fmt.Sprintf("%s.%s", domain, service)
			var spiffeURI string
			if spiffe || spiffeTrustDomain != "" {
				if spiffeTrustDomain != "" {
					spiffeURI = fmt.Sprintf("spiffe://%s/ns/%s/ra/%s", spiffeTrustDomain, roleDomain, roleName)
				} else {
					spiffeURI = fmt.Sprintf("spiffe://%s/ra/%s", roleDomain, roleName)
				}
			}
			subj := newCSRSubject(fmt.Sprintf("%s:role.%s", roleDomain, roleName), subjC, subjP, subjO, subjOU)
			csrPEM, err := generateRoleCSR(signer, subj, host, principal, dnsDomain, ip, spiffeURI)
			if err != nil {
				return err
			}
			if csrOnly {
				_, err := fmt.Fprint(cmd.OutOrStdout(), csrPEM)
				return err
			}

			req := &zts.RoleCertificateRequest{
				Csr:               csrPEM,
				ExpiryTime:        int64(expiryTime),
				ProxyForPrincipal: zts.EntityName(proxyForPrincipal),
			}
			if signerKeyID != "" {
				req.X509CertSignerKeyId = zts.SimpleName(signerKeyID)
			}
			if oldRoleCertPath != "" {
				prev, err := certFromFile(oldRoleCertPath)
				if err != nil {
					return fmt.Errorf("read --old-role-cert: %w", err)
				}
				req.PrevCertNotBefore = &rdl.Timestamp{Time: prev.NotBefore}
				req.PrevCertNotAfter = &rdl.Timestamp{Time: prev.NotAfter}
			}

			zc, err := opts.ZTSClient()
			if err != nil {
				return err
			}
			resp, err := zc.PostRoleCertificateRequestExt(req)
			if err != nil {
				return cliopts.WrapErr(err)
			}

			certificate := resp.X509Certificate

			if concatIntermediateCert {
				_, caCert := pem.Decode([]byte(certificate))
				if len(caCert) == 0 {
					intermediateCertBundle, err := zc.GetCertificateAuthorityBundle(zts.SimpleName(caCertBundleName))
					if err != nil || intermediateCertBundle == nil || intermediateCertBundle.Certs == "" {
						return fmt.Errorf("GetCertificateAuthorityBundle failed for role certificate, err: %v", err)
					}
					intermediateCerts := intermediateCertBundle.Certs
					certificate += intermediateCerts
				}
			}

			return writePEM(cmd, outPath, certificate)
		},
	}
	// flagDefaults seeds the flag help text with any build-time (ldflags -X)
	// overrides so `--help` reflects the same defaults resolveCertificateDefaults
	// applies at run time. Errors here (a malformed built-in override) surface
	// again, properly, when the command actually runs.
	flagDefaults, _ := buildCertificateDefaults(roleCertKind)
	cmd.Flags().StringVar(&roleDomain, "role-domain", "", "role domain (required)")
	cmd.Flags().StringVar(&roleName, "role-name", "", "role name in --role-domain (required)")
	cmd.Flags().StringVar(&service, "service", "", "caller service name (default: extracted from context cert CN)")
	cmd.Flags().StringVar(&roleKeyFile, "role-key-file", "", "role cert private key file (default: context service key)")
	cmd.Flags().StringVar(&dnsDomain, "dns-domain", flagDefaults.dnsDomain, "DNS domain suffix to include in the CSR SAN (required unless configured)")
	cmd.Flags().StringVar(&subjC, "subj-c", flagDefaults.subjectCountry, "CSR Subject Country")
	cmd.Flags().StringVar(&subjP, "subj-p", flagDefaults.subjectProvince, "CSR Subject Province")
	cmd.Flags().StringVar(&subjO, "subj-o", flagDefaults.subjectOrganization, "CSR Subject Organization")
	cmd.Flags().StringVar(&subjOU, "subj-ou", flagDefaults.subjectOrganizationalUnit, "CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&spiffe, "spiffe", flagDefaults.spiffe, "include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&spiffeTrustDomain, "spiffe-trust-domain", flagDefaults.spiffeTrustDomain, "SPIFFE trust domain")
	cmd.Flags().StringVar(&ip, "ip", flagDefaults.ip, "IP address to include in CSR SAN")
	cmd.Flags().StringVar(&oldRoleCertPath, "old-role-cert", "", "path to previous role cert PEM (sent as PrevCertNotBefore/NotAfter)")
	cmd.Flags().BoolVar(&csrOnly, "csr", false, "print the generated CSR and exit")
	cmd.Flags().IntVar(&expiryTime, "expiry-time", flagDefaults.expiryTimeMinutes, "requested certificate lifetime in minutes (0 = server default)")
	cmd.Flags().StringVar(&outPath, "out", "", "path to write the signed cert (default: stdout)")
	cmd.Flags().StringVar(&proxyForPrincipal, "proxy-for-principal", "", "issue cert proxied on behalf of this principal")
	cmd.Flags().BoolVar(&concatIntermediateCert, "concat-intermediate-cert", flagDefaults.concatIntermediateCert, "append a CA bundle when the response does not include a certificate chain")
	cmd.Flags().StringVar(&caCertBundleName, "cacert-bundle-name", flagDefaults.caCertBundleName, "CA certificate bundle name used with --concat-intermediate-cert")
	cmd.Flags().StringVar(&signerKeyID, "signer-key-id", flagDefaults.signerKeyID, "ZTS certificate signer key id")
	return cmd
}

func writePEM(cmd *cobra.Command, outPath, pem string) error {
	if outPath == "" || outPath == "-" {
		_, err := fmt.Fprint(cmd.OutOrStdout(), pem)
		return err
	}
	return os.WriteFile(outPath, []byte(pem), 0o644)
}

func generateRoleCSR(signer *csrSigner, subj pkix.Name, host, principal, dnsDomain, ip, spiffeURI string) (string, error) {
	tmpl := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: signer.algorithm,
	}
	if host != "" {
		tmpl.DNSNames = []string{host}
	}
	if ip != "" {
		tmpl.IPAddresses = []net.IP{net.ParseIP(ip)}
	}
	tmpl.EmailAddresses = []string{fmt.Sprintf("%s@%s", principal, dnsDomain)}
	if spiffeURI != "" {
		if u, err := url.Parse(spiffeURI); err == nil {
			tmpl.URIs = append(tmpl.URIs, u)
		}
	}
	if u, err := url.Parse(fmt.Sprintf("athenz://principal/%s", principal)); err == nil {
		tmpl.URIs = append(tmpl.URIs, u)
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, &tmpl, signer.key)
	if err != nil {
		return "", fmt.Errorf("create CSR: %w", err)
	}
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}); err != nil {
		return "", fmt.Errorf("encode CSR: %w", err)
	}
	return buf.String(), nil
}

func certFromFile(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func extractServiceFromCertFile(path string) (domain, service string, err error) {
	cert, err := certFromFile(path)
	if err != nil {
		return "", "", err
	}
	cn := cert.Subject.CommonName
	idx := strings.LastIndex(cn, ".")
	if idx < 0 {
		return "", "", fmt.Errorf("cannot split domain/service from CN %q", cn)
	}
	return cn[:idx], cn[idx+1:], nil
}
