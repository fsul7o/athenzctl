package issue

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// newServiceCert issues or refreshes a service X.509 identity certificate.
// Flags mirror `zts-svccert` (with `--` prefix); ZTS URL and mTLS credentials
// come from the athenzctl context.
func newServiceCert(opts *cliopts.Options) *cobra.Command {
	var (
		service                     string
		provider                    string
		instance                    string
		privateKeyPath              string
		dnsDomain                   string
		subjC, subjP, subjO, subjOU string
		spiffe                      bool
		spiffeTrustDomain           string
		ip                          string
		attestationDataFile         string
		signerKeyID                 string
		expiryTime                  int
		outPath                     string
		signerOutPath               string
		csrOnly                     bool
		useInstanceRegisterToken    bool
		concatIntermediateCert      bool
	)
	cmd := &cobra.Command{
		Use:   "servicecert",
		Short: "Issue or refresh a service X.509 identity certificate",
		Long: `Issue or refresh a service X.509 identity certificate via ZTS.

Modes (mirrors zts-svccert):
  - CSR only:   --csr --private-key ... --dns-domain ...
  - Register:   --provider ... --instance ... --private-key ... --dns-domain ...
                (attestation from --attestation-data file, or
                 --use-instance-register-token to fetch one automatically)
  - Refresh:    --private-key ... --dns-domain ... (no --provider)

ZTS URL and mTLS credentials always come from the athenzctl context.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			defaults, err := resolveCertificateDefaults(cmd, opts, serviceCertKind)
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

			domain, err := opts.RequireDomain()
			if err != nil {
				return err
			}
			if service == "" {
				return errors.New("issue servicecert requires --service")
			}
			if privateKeyPath == "" {
				return errors.New("issue servicecert requires --private-key")
			}
			if dnsDomain == "" {
				return errors.New("issue servicecert requires --dns-domain or a servicecert default")
			}
			keyBytes, err := os.ReadFile(privateKeyPath)
			if err != nil {
				if !useInstanceRegisterToken || !os.IsNotExist(err) {
					return fmt.Errorf("read private key: %w", err)
				}
				keyBytes, err = generateRSAPrivateKey(privateKeyPath)
				if err != nil {
					return fmt.Errorf("generate private key: %w", err)
				}
			}
			signer, err := newCSRSigner(keyBytes)
			if err != nil {
				return fmt.Errorf("load private key: %w", err)
			}

			hyphenDomain := strings.ReplaceAll(domain, ".", "-")
			host := fmt.Sprintf("%s.%s.%s", service, hyphenDomain, dnsDomain)
			commonName := fmt.Sprintf("%s.%s", domain, service)
			var instanceURI string
			if instance != "" {
				uriProvider := "zts"
				if provider != "" {
					uriProvider = provider
				}
				instanceURI = fmt.Sprintf("athenz://instanceid/%s/%s", uriProvider, instance)
			}
			var spiffeURI string
			if spiffe || spiffeTrustDomain != "" {
				if spiffeTrustDomain != "" {
					spiffeURI = fmt.Sprintf("spiffe://%s/ns/default/sa/%s", spiffeTrustDomain, commonName)
				} else {
					spiffeURI = fmt.Sprintf("spiffe://%s/sa/%s", domain, service)
				}
			}
			subj := newCSRSubject(commonName, subjC, subjP, subjO, subjOU)
			csrPEM, err := generateCSR(signer, subj, host, instanceURI, ip, spiffeURI)
			if err != nil {
				return err
			}
			if csrOnly {
				_, err := fmt.Fprint(cmd.OutOrStdout(), csrPEM)
				return err
			}

			zc, err := opts.ZTSClient()
			if err != nil {
				return err
			}

			var certificate, signerCert string
			if provider != "" {
				if instance == "" {
					return errors.New("--provider requires --instance")
				}
				attestationData, err := resolveAttestationData(zc, provider, domain, service, instance, attestationDataFile, useInstanceRegisterToken)
				if err != nil {
					return err
				}
				req := &zts.InstanceRegisterInformation{
					Provider:        zts.ServiceName(provider),
					Domain:          zts.DomainName(domain),
					Service:         zts.SimpleName(service),
					AttestationData: attestationData,
					Csr:             csrPEM,
				}
				if signerKeyID != "" {
					req.X509CertSignerKeyId = zts.SimpleName(signerKeyID)
				}
				ident, _, err := zc.PostInstanceRegisterInformation(req)
				if err != nil {
					return cliopts.WrapErr(err)
				}
				certificate = ident.X509Certificate
				signerCert = ident.X509CertificateSigner
			} else {
				expiry32 := int32(expiryTime)
				req := &zts.InstanceRefreshRequest{
					Csr:        csrPEM,
					ExpiryTime: &expiry32,
				}
				if signerKeyID != "" {
					req.X509CertSignerKeyId = zts.SimpleName(signerKeyID)
				}
				ident, err := zc.PostInstanceRefreshRequest(zts.CompoundName(domain), zts.SimpleName(service), req)
				if err != nil {
					return cliopts.WrapErr(err)
				}
				certificate = ident.Certificate
				if concatIntermediateCert {
					certificate += ident.CaCertBundle
				}
				signerCert = ident.CaCertBundle
			}

			if signerOutPath != "" && signerCert != "" {
				if err := os.WriteFile(signerOutPath, []byte(signerCert), 0o644); err != nil {
					return fmt.Errorf("write signer cert: %w", err)
				}
			}
			return writePEM(cmd, outPath, certificate)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "Athenz provider service name (initial registration only)")
	cmd.Flags().StringVar(&instance, "instance", "", "instance ID")
	cmd.Flags().StringVar(&privateKeyPath, "private-key", "", "path to the service private key PEM (required)")
	cmd.Flags().StringVar(&dnsDomain, "dns-domain", "", "DNS domain suffix to include in the CSR SAN (required unless configured)")
	cmd.Flags().StringVar(&subjC, "subj-c", "US", "CSR Subject Country")
	cmd.Flags().StringVar(&subjP, "subj-p", "", "CSR Subject Province")
	cmd.Flags().StringVar(&subjO, "subj-o", "Oath Inc.", "CSR Subject Organization")
	cmd.Flags().StringVar(&subjOU, "subj-ou", "Athenz", "CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&spiffe, "spiffe", true, "include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&spiffeTrustDomain, "spiffe-trust-domain", "", "SPIFFE trust domain")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address to include in CSR SAN")
	cmd.Flags().StringVar(&attestationDataFile, "attestation-data", "", "attestation data file (for --provider registration)")
	cmd.Flags().StringVar(&signerKeyID, "signer-key-id", "", "ZTS certificate signer key id")
	cmd.Flags().IntVar(&expiryTime, "expiry-time", 0, "requested certificate lifetime in minutes (0 = server default)")
	cmd.Flags().StringVar(&outPath, "out", "", "path to write the signed cert (default: stdout)")
	cmd.Flags().StringVar(&signerOutPath, "signer-cert-out", "", "path to write the signer CA cert if returned")
	cmd.Flags().BoolVar(&csrOnly, "csr", false, "print the generated CSR and exit")
	cmd.Flags().BoolVar(&useInstanceRegisterToken, "use-instance-register-token", false, "fetch an instance register token via the current context and use it as attestation data")
	cmd.Flags().BoolVar(&concatIntermediateCert, "concat-intermediate-cert", false, "append the returned intermediate CA bundle to the certificate")
	return cmd
}

// resolveAttestationData picks the attestation data source in this order:
//  1. --attestation-data file, if given
//  2. --use-instance-register-token: fetch a fresh register token via ZTS
//  3. empty string (provider must accept unsigned/mTLS-authenticated register)
func resolveAttestationData(zc *zts.ZTSClient, provider, domain, service, instance, file string, useToken bool) (string, error) {
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read attestation data: %w", err)
		}
		return string(b), nil
	}
	if useToken {
		tok, err := zc.GetInstanceRegisterToken(
			zts.ServiceName(provider),
			zts.DomainName(domain),
			zts.SimpleName(service),
			zts.PathElement(instance),
		)
		if err != nil {
			return "", cliopts.WrapErr(err)
		}
		return tok.AttestationData, nil
	}
	return "", nil
}
