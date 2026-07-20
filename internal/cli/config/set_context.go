package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func newSetContext(opts *Options) *cobra.Command {
	var (
		zmsURL                string
		ztsURL                string
		cert                  string
		key                   string
		caCert                string
		zmsServerName         string
		ztsServerName         string
		insecureSkipTLSVerify bool
		proxyURL              string
		authMode              string
		// exec fields
		execCommand  string
		execArgs     []string
		execEnv      []string
		execCertPath string
		execKeyPath  string
		// ntoken-auth fields
		ntokenAuthDomain     string
		ntokenAuthService    string
		ntokenAuthPrivateKey string
		ntokenAuthKeyID      string
		ntokenAuthHeader     string
		// copperargos-auth fields
		copperArgosAuthDomain          string
		copperArgosAuthService         string
		copperArgosAuthProvider        string
		copperArgosAuthInstance        string
		copperArgosAuthAttestationData string
		// shared by ntoken-auth/copperargos-auth (a context only uses one auth-mode at a time)
		authCacheDir           string
		serviceCertSubjC       string
		serviceCertSubjP       string
		serviceCertSubjO       string
		serviceCertSubjOU      string
		serviceCertSpiffe      bool
		serviceCertTrustDomain string
		serviceCertDNSDomain   string
		serviceCertConcat      bool
		serviceCertExpiryTime  int
		serviceCertIP          string
		serviceCertSignerKeyID string
		roleCertSubjC          string
		roleCertSubjP          string
		roleCertSubjO          string
		roleCertSubjOU         string
		roleCertSpiffe         bool
		roleCertTrustDomain    string
		roleCertDNSDomain      string
		roleCertConcat         bool
		roleCertBundleName     string
		roleCertExpiryTime     int
		roleCertIP             string
		roleCertSignerKeyID    string
	)
	cmd := &cobra.Command{
		Use:   "set-context NAME",
		Short: "Create or update a context in the athenzctl config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, path, err := loadConfig(opts)
			if err != nil {
				return err
			}

			existing := c.Find(name)
			ctx := cfg.Context{Name: name}
			if existing != nil {
				ctx = *existing
			}
			if cmd.Flags().Changed("zms-url") {
				ctx.ZMSURL = zmsURL
			}
			if cmd.Flags().Changed("zts-url") {
				ctx.ZTSURL = ztsURL
			}
			if cmd.Flags().Changed("cert") {
				ctx.Cert = cert
			}
			if cmd.Flags().Changed("key") {
				ctx.Key = key
			}
			if cmd.Flags().Changed("ca-cert") {
				ctx.CACert = caCert
			}
			if cmd.Flags().Changed("zms-server-name") {
				ctx.ZMSServerName = zmsServerName
			}
			if cmd.Flags().Changed("zts-server-name") {
				ctx.ZTSServerName = ztsServerName
			}
			if cmd.Flags().Changed("insecure-skip-tls-verify") {
				ctx.InsecureSkipTLSVerify = insecureSkipTLSVerify
			}
			if cmd.Flags().Changed("proxy") {
				ctx.ProxyURL = proxyURL
			}
			if cmd.Flags().Changed("auth-mode") {
				ctx.AuthMode = authMode
			}
			if cmd.Flags().Changed("auth-cache-dir") {
				ctx.CacheDir = authCacheDir
			}

			serviceCertDefaults := certificateDefaultValues{
				subjectCountry:            serviceCertSubjC,
				subjectProvince:           serviceCertSubjP,
				subjectOrganization:       serviceCertSubjO,
				subjectOrganizationalUnit: serviceCertSubjOU,
				spiffe:                    serviceCertSpiffe,
				spiffeTrustDomain:         serviceCertTrustDomain,
				dnsDomain:                 serviceCertDNSDomain,
				concatIntermediateCert:    serviceCertConcat,
				expiryTimeMinutes:         serviceCertExpiryTime,
				ip:                        serviceCertIP,
				signerKeyID:               serviceCertSignerKeyID,
			}
			roleCertDefaults := certificateDefaultValues{
				subjectCountry:            roleCertSubjC,
				subjectProvince:           roleCertSubjP,
				subjectOrganization:       roleCertSubjO,
				subjectOrganizationalUnit: roleCertSubjOU,
				spiffe:                    roleCertSpiffe,
				spiffeTrustDomain:         roleCertTrustDomain,
				dnsDomain:                 roleCertDNSDomain,
				concatIntermediateCert:    roleCertConcat,
				expiryTimeMinutes:         roleCertExpiryTime,
				ip:                        roleCertIP,
				signerKeyID:               roleCertSignerKeyID,
			}
			serviceCertChanged := certificateDefaultFlagsChanged(cmd, "servicecert")
			roleCertChanged := certificateDefaultFlagsChanged(cmd, "rolecert") || cmd.Flags().Changed("rolecert-cacert-bundle-name")
			if serviceCertChanged || roleCertChanged {
				if ctx.IssueDefaults == nil {
					ctx.IssueDefaults = &cfg.IssueDefaults{}
				}
			}
			if serviceCertChanged {
				if ctx.IssueDefaults.ServiceCert == nil {
					ctx.IssueDefaults.ServiceCert = &cfg.CertificateDefaults{}
				}
				applyCertificateDefaultFlags(cmd, ctx.IssueDefaults.ServiceCert, "servicecert", serviceCertDefaults)
			}
			if roleCertChanged {
				if ctx.IssueDefaults.RoleCert == nil {
					ctx.IssueDefaults.RoleCert = &cfg.CertificateDefaults{}
				}
				defaults := ctx.IssueDefaults.RoleCert
				applyCertificateDefaultFlags(cmd, defaults, "rolecert", roleCertDefaults)
				if cmd.Flags().Changed("rolecert-cacert-bundle-name") {
					defaults.CACertBundleName = roleCertBundleName
				}
			}

			// Ensure ctx.Exec exists if any exec-*  flag was set.
			execChanged := false
			for _, f := range []string{"exec-command", "exec-arg", "exec-env", "exec-cert-path", "exec-key-path"} {
				if cmd.Flags().Changed(f) {
					execChanged = true
					break
				}
			}
			if execChanged && ctx.Exec == nil {
				ctx.Exec = &cfg.ExecConfig{}
			}
			if ctx.Exec != nil {
				if cmd.Flags().Changed("exec-command") {
					ctx.Exec.Command = execCommand
				}
				if cmd.Flags().Changed("exec-arg") {
					ctx.Exec.Args = execArgs
				}
				if cmd.Flags().Changed("exec-env") {
					env := make(map[string]string, len(execEnv))
					for _, kv := range execEnv {
						k, v, ok := strings.Cut(kv, "=")
						if !ok {
							return fmt.Errorf("--exec-env %q: expected KEY=VALUE", kv)
						}
						env[k] = v
					}
					ctx.Exec.Env = env
				}
				if cmd.Flags().Changed("exec-cert-path") {
					ctx.Exec.CertPath = execCertPath
				}
				if cmd.Flags().Changed("exec-key-path") {
					ctx.Exec.KeyPath = execKeyPath
				}
			}

			ntokenAuthChanged := anyFlagChanged(cmd,
				"ntoken-auth-domain", "ntoken-auth-service", "ntoken-auth-private-key",
				"ntoken-auth-key-id", "ntoken-auth-hdr")
			if ntokenAuthChanged && ctx.NTokenAuth == nil {
				ctx.NTokenAuth = &cfg.NTokenAuthConfig{}
			}
			if ctx.NTokenAuth != nil {
				if cmd.Flags().Changed("ntoken-auth-domain") {
					ctx.NTokenAuth.Domain = ntokenAuthDomain
				}
				if cmd.Flags().Changed("ntoken-auth-service") {
					ctx.NTokenAuth.Service = ntokenAuthService
				}
				if cmd.Flags().Changed("ntoken-auth-private-key") {
					ctx.NTokenAuth.PrivateKeyPath = ntokenAuthPrivateKey
				}
				if cmd.Flags().Changed("ntoken-auth-key-id") {
					ctx.NTokenAuth.KeyID = ntokenAuthKeyID
				}
				if cmd.Flags().Changed("ntoken-auth-hdr") {
					ctx.NTokenAuth.Header = ntokenAuthHeader
				}
			}

			copperArgosAuthChanged := anyFlagChanged(cmd,
				"copperargos-auth-domain", "copperargos-auth-service", "copperargos-auth-provider",
				"copperargos-auth-instance", "copperargos-auth-attestation-data")
			if copperArgosAuthChanged && ctx.CopperArgosAuth == nil {
				ctx.CopperArgosAuth = &cfg.CopperArgosAuthConfig{}
			}
			if ctx.CopperArgosAuth != nil {
				if cmd.Flags().Changed("copperargos-auth-domain") {
					ctx.CopperArgosAuth.Domain = copperArgosAuthDomain
				}
				if cmd.Flags().Changed("copperargos-auth-service") {
					ctx.CopperArgosAuth.Service = copperArgosAuthService
				}
				if cmd.Flags().Changed("copperargos-auth-provider") {
					ctx.CopperArgosAuth.Provider = copperArgosAuthProvider
				}
				if cmd.Flags().Changed("copperargos-auth-instance") {
					ctx.CopperArgosAuth.Instance = copperArgosAuthInstance
				}
				if cmd.Flags().Changed("copperargos-auth-attestation-data") {
					ctx.CopperArgosAuth.AttestationDataPath = copperArgosAuthAttestationData
				}
			}

			c.Upsert(ctx)
			if err := cfg.Save(path, c); err != nil {
				return err
			}
			verb := "created"
			if existing != nil {
				verb = "updated"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Context %q %s in %s\n", name, verb, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&zmsURL, "zms-url", "", "ZMS server URL (e.g. https://zms.example.com:4443/zms/v1)")
	cmd.Flags().StringVar(&ztsURL, "zts-url", "", "ZTS server URL (e.g. https://zts.example.com:4443/zts/v1)")
	cmd.Flags().StringVar(&cert, "cert", "", "path to client certificate (PEM) for mTLS")
	cmd.Flags().StringVar(&key, "key", "", "path to client private key (PEM) for mTLS")
	cmd.Flags().StringVar(&caCert, "ca-cert", "", "path to CA bundle (PEM) for verifying the server")
	cmd.Flags().StringVar(&zmsServerName, "zms-server-name", "", "TLS ServerName override for ZMS (SNI + hostname verification)")
	cmd.Flags().StringVar(&ztsServerName, "zts-server-name", "", "TLS ServerName override for ZTS (SNI + hostname verification)")
	cmd.Flags().BoolVarP(&insecureSkipTLSVerify, "insecure-skip-tls-verify", "k", false, "disable TLS certificate and hostname verification")
	cmd.Flags().StringVarP(&proxyURL, "proxy", "s", "", "proxy URL (host:port for SOCKS5, or socks5/http/https URL)")
	cmd.Flags().StringVar(&authMode, "auth-mode", "", "authentication mode: \"\" or \"mtls\" (default), \"exec\" (obtain the client cert by execing an external command that places it at a known path), \"ntoken\" (mint a fresh cert on every invocation via a ZMS-registered key pair), \"copperargos\" (register/self-refresh a cert via a prepared attestation-data file)")
	// exec flags
	cmd.Flags().StringVar(&execCommand, "exec-command", "", "exec: path to the external command that places a fresh cert/key at exec-cert-path/exec-key-path")
	cmd.Flags().StringArrayVar(&execArgs, "exec-arg", nil, "exec: argument to pass to the exec command (repeatable, replaces the full list when set)")
	cmd.Flags().StringArrayVar(&execEnv, "exec-env", nil, "exec: KEY=VALUE environment variable to set for the exec command (repeatable, replaces the full map when set)")
	cmd.Flags().StringVar(&execCertPath, "exec-cert-path", "", "exec: path the exec command writes the cert PEM to, read back after it exits")
	cmd.Flags().StringVar(&execKeyPath, "exec-key-path", "", "exec: path the exec command writes the key PEM to, read back after it exits")
	cmd.Flags().StringVar(&ntokenAuthDomain, "ntoken-auth-domain", "", "ntoken: domain of the service identity to mint a certificate for")
	cmd.Flags().StringVar(&ntokenAuthService, "ntoken-auth-service", "", "ntoken: name of the service identity to mint a certificate for")
	cmd.Flags().StringVar(&ntokenAuthPrivateKey, "ntoken-auth-private-key", "", "ntoken: path to the ZMS-registered private key PEM")
	cmd.Flags().StringVar(&ntokenAuthKeyID, "ntoken-auth-key-id", "", "ntoken: ZMS public key version for --ntoken-auth-private-key")
	cmd.Flags().StringVar(&ntokenAuthHeader, "ntoken-auth-hdr", "", "ntoken: HTTP header the signed NToken is sent under (default: \"Athenz-Principal-Auth\", mirrors zts-svccert -hdr)")
	cmd.Flags().StringVar(&copperArgosAuthDomain, "copperargos-auth-domain", "", "copperargos: domain of the service identity to register")
	cmd.Flags().StringVar(&copperArgosAuthService, "copperargos-auth-service", "", "copperargos: name of the service identity to register")
	cmd.Flags().StringVar(&copperArgosAuthProvider, "copperargos-auth-provider", "", "copperargos: Athenz provider service name")
	cmd.Flags().StringVar(&copperArgosAuthInstance, "copperargos-auth-instance", "", "copperargos: instance ID to register")
	cmd.Flags().StringVar(&copperArgosAuthAttestationData, "copperargos-auth-attestation-data", "", "copperargos: path to a prepared attestation-data file (e.g. from `issue instance-register-token --out`)")
	cmd.Flags().StringVar(&authCacheDir, "auth-cache-dir", "", "ntoken/copperargos: cache directory for the minted/issued cert+key (default: ~/.athenzctl/cache/<context>/<mode>)")
	cmd.Flags().StringVar(&serviceCertSubjC, "servicecert-subj-c", "", "servicecert default CSR Subject Country")
	cmd.Flags().StringVar(&serviceCertSubjP, "servicecert-subj-p", "", "servicecert default CSR Subject Province")
	cmd.Flags().StringVar(&serviceCertSubjO, "servicecert-subj-o", "", "servicecert default CSR Subject Organization")
	cmd.Flags().StringVar(&serviceCertSubjOU, "servicecert-subj-ou", "", "servicecert default CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&serviceCertSpiffe, "servicecert-spiffe", true, "servicecert default: include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&serviceCertTrustDomain, "servicecert-spiffe-trust-domain", "", "servicecert default SPIFFE trust domain")
	cmd.Flags().StringVar(&serviceCertDNSDomain, "servicecert-dns-domain", "", "servicecert default DNS domain suffix")
	cmd.Flags().BoolVar(&serviceCertConcat, "servicecert-concat-intermediate-cert", false, "servicecert default: append the returned intermediate CA bundle to the certificate")
	cmd.Flags().IntVar(&serviceCertExpiryTime, "servicecert-expiry-time", 0, "servicecert default requested certificate lifetime in minutes (0 = server default)")
	cmd.Flags().StringVar(&serviceCertIP, "servicecert-ip", "", "servicecert default IP address to include in CSR SAN")
	cmd.Flags().StringVar(&serviceCertSignerKeyID, "servicecert-signer-key-id", "", "servicecert default ZTS certificate signer key id")
	cmd.Flags().StringVar(&roleCertSubjC, "rolecert-subj-c", "", "rolecert default CSR Subject Country")
	cmd.Flags().StringVar(&roleCertSubjP, "rolecert-subj-p", "", "rolecert default CSR Subject Province")
	cmd.Flags().StringVar(&roleCertSubjO, "rolecert-subj-o", "", "rolecert default CSR Subject Organization")
	cmd.Flags().StringVar(&roleCertSubjOU, "rolecert-subj-ou", "", "rolecert default CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&roleCertSpiffe, "rolecert-spiffe", true, "rolecert default: include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&roleCertTrustDomain, "rolecert-spiffe-trust-domain", "", "rolecert default SPIFFE trust domain")
	cmd.Flags().StringVar(&roleCertDNSDomain, "rolecert-dns-domain", "", "rolecert default DNS domain suffix")
	cmd.Flags().BoolVar(&roleCertConcat, "rolecert-concat-intermediate-cert", false, "rolecert default: append a CA bundle when the response does not include a certificate chain")
	cmd.Flags().StringVar(&roleCertBundleName, "rolecert-cacert-bundle-name", "", "rolecert default CA certificate bundle name used with --concat-intermediate-cert")
	cmd.Flags().IntVar(&roleCertExpiryTime, "rolecert-expiry-time", 0, "rolecert default requested certificate lifetime in minutes (0 = server default)")
	cmd.Flags().StringVar(&roleCertIP, "rolecert-ip", "", "rolecert default IP address to include in CSR SAN")
	cmd.Flags().StringVar(&roleCertSignerKeyID, "rolecert-signer-key-id", "", "rolecert default ZTS certificate signer key id")
	return cmd
}

func anyFlagChanged(cmd *cobra.Command, names ...string) bool {
	for _, name := range names {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

type certificateDefaultValues struct {
	subjectCountry            string
	subjectProvince           string
	subjectOrganization       string
	subjectOrganizationalUnit string
	spiffe                    bool
	spiffeTrustDomain         string
	dnsDomain                 string
	concatIntermediateCert    bool
	expiryTimeMinutes         int
	ip                        string
	signerKeyID               string
}

func certificateDefaultFlagsChanged(cmd *cobra.Command, prefix string) bool {
	for _, suffix := range []string{
		"subj-c", "subj-p", "subj-o", "subj-ou", "spiffe", "spiffe-trust-domain",
		"dns-domain", "concat-intermediate-cert", "expiry-time", "ip", "signer-key-id",
	} {
		if cmd.Flags().Changed(prefix + "-" + suffix) {
			return true
		}
	}
	return false
}

func applyCertificateDefaultFlags(cmd *cobra.Command, defaults *cfg.CertificateDefaults, prefix string, values certificateDefaultValues) {
	if cmd.Flags().Changed(prefix + "-subj-c") {
		defaults.SubjectCountry = values.subjectCountry
	}
	if cmd.Flags().Changed(prefix + "-subj-p") {
		defaults.SubjectProvince = values.subjectProvince
	}
	if cmd.Flags().Changed(prefix + "-subj-o") {
		defaults.SubjectOrganization = values.subjectOrganization
	}
	if cmd.Flags().Changed(prefix + "-subj-ou") {
		defaults.SubjectOrganizationalUnit = values.subjectOrganizationalUnit
	}
	if cmd.Flags().Changed(prefix + "-spiffe") {
		value := values.spiffe
		defaults.Spiffe = &value
	}
	if cmd.Flags().Changed(prefix + "-spiffe-trust-domain") {
		defaults.SpiffeTrustDomain = values.spiffeTrustDomain
	}
	if cmd.Flags().Changed(prefix + "-dns-domain") {
		defaults.DNSDomain = values.dnsDomain
	}
	if cmd.Flags().Changed(prefix + "-concat-intermediate-cert") {
		value := values.concatIntermediateCert
		defaults.ConcatIntermediateCert = &value
	}
	if cmd.Flags().Changed(prefix + "-expiry-time") {
		defaults.ExpiryTimeMinutes = values.expiryTimeMinutes
	}
	if cmd.Flags().Changed(prefix + "-ip") {
		defaults.IP = values.ip
	}
	if cmd.Flags().Changed(prefix + "-signer-key-id") {
		defaults.SignerKeyID = values.signerKeyID
	}
}
