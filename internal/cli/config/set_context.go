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
		execCommand            string
		execArgs               []string
		execEnv                []string
		execCertPath           string
		execKeyPath            string
		serviceCertSubjC       string
		serviceCertSubjP       string
		serviceCertSubjO       string
		serviceCertSubjOU      string
		serviceCertSpiffe      bool
		serviceCertTrustDomain string
		serviceCertDNSDomain   string
		roleCertSubjC          string
		roleCertSubjP          string
		roleCertSubjO          string
		roleCertSubjOU         string
		roleCertSpiffe         bool
		roleCertTrustDomain    string
		roleCertDNSDomain      string
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

			serviceCertChanged := anyFlagChanged(cmd,
				"servicecert-subj-c", "servicecert-subj-p", "servicecert-subj-o", "servicecert-subj-ou",
				"servicecert-spiffe", "servicecert-spiffe-trust-domain", "servicecert-dns-domain")
			roleCertChanged := anyFlagChanged(cmd,
				"rolecert-subj-c", "rolecert-subj-p", "rolecert-subj-o", "rolecert-subj-ou",
				"rolecert-spiffe", "rolecert-spiffe-trust-domain", "rolecert-dns-domain")
			if serviceCertChanged || roleCertChanged {
				if ctx.IssueDefaults == nil {
					ctx.IssueDefaults = &cfg.IssueDefaults{}
				}
			}
			if serviceCertChanged {
				if ctx.IssueDefaults.ServiceCert == nil {
					ctx.IssueDefaults.ServiceCert = &cfg.CertificateDefaults{}
				}
				defaults := ctx.IssueDefaults.ServiceCert
				if cmd.Flags().Changed("servicecert-subj-c") {
					defaults.SubjectCountry = serviceCertSubjC
				}
				if cmd.Flags().Changed("servicecert-subj-p") {
					defaults.SubjectProvince = serviceCertSubjP
				}
				if cmd.Flags().Changed("servicecert-subj-o") {
					defaults.SubjectOrganization = serviceCertSubjO
				}
				if cmd.Flags().Changed("servicecert-subj-ou") {
					defaults.SubjectOrganizationalUnit = serviceCertSubjOU
				}
				if cmd.Flags().Changed("servicecert-spiffe") {
					value := serviceCertSpiffe
					defaults.Spiffe = &value
				}
				if cmd.Flags().Changed("servicecert-spiffe-trust-domain") {
					defaults.SpiffeTrustDomain = serviceCertTrustDomain
				}
				if cmd.Flags().Changed("servicecert-dns-domain") {
					defaults.DNSDomain = serviceCertDNSDomain
				}
			}
			if roleCertChanged {
				if ctx.IssueDefaults.RoleCert == nil {
					ctx.IssueDefaults.RoleCert = &cfg.CertificateDefaults{}
				}
				defaults := ctx.IssueDefaults.RoleCert
				if cmd.Flags().Changed("rolecert-subj-c") {
					defaults.SubjectCountry = roleCertSubjC
				}
				if cmd.Flags().Changed("rolecert-subj-p") {
					defaults.SubjectProvince = roleCertSubjP
				}
				if cmd.Flags().Changed("rolecert-subj-o") {
					defaults.SubjectOrganization = roleCertSubjO
				}
				if cmd.Flags().Changed("rolecert-subj-ou") {
					defaults.SubjectOrganizationalUnit = roleCertSubjOU
				}
				if cmd.Flags().Changed("rolecert-spiffe") {
					value := roleCertSpiffe
					defaults.Spiffe = &value
				}
				if cmd.Flags().Changed("rolecert-spiffe-trust-domain") {
					defaults.SpiffeTrustDomain = roleCertTrustDomain
				}
				if cmd.Flags().Changed("rolecert-dns-domain") {
					defaults.DNSDomain = roleCertDNSDomain
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
	cmd.Flags().StringVar(&authMode, "auth-mode", "", "authentication mode: \"\" or \"mtls\" (default), \"exec\" (obtain the client cert by execing an external command that places it at a known path)")
	// exec flags
	cmd.Flags().StringVar(&execCommand, "exec-command", "", "exec: path to the external command that places a fresh cert/key at exec-cert-path/exec-key-path")
	cmd.Flags().StringArrayVar(&execArgs, "exec-arg", nil, "exec: argument to pass to the exec command (repeatable, replaces the full list when set)")
	cmd.Flags().StringArrayVar(&execEnv, "exec-env", nil, "exec: KEY=VALUE environment variable to set for the exec command (repeatable, replaces the full map when set)")
	cmd.Flags().StringVar(&execCertPath, "exec-cert-path", "", "exec: path the exec command writes the cert PEM to, read back after it exits")
	cmd.Flags().StringVar(&execKeyPath, "exec-key-path", "", "exec: path the exec command writes the key PEM to, read back after it exits")
	cmd.Flags().StringVar(&serviceCertSubjC, "servicecert-subj-c", "", "servicecert default CSR Subject Country")
	cmd.Flags().StringVar(&serviceCertSubjP, "servicecert-subj-p", "", "servicecert default CSR Subject Province")
	cmd.Flags().StringVar(&serviceCertSubjO, "servicecert-subj-o", "", "servicecert default CSR Subject Organization")
	cmd.Flags().StringVar(&serviceCertSubjOU, "servicecert-subj-ou", "", "servicecert default CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&serviceCertSpiffe, "servicecert-spiffe", true, "servicecert default: include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&serviceCertTrustDomain, "servicecert-spiffe-trust-domain", "", "servicecert default SPIFFE trust domain")
	cmd.Flags().StringVar(&serviceCertDNSDomain, "servicecert-dns-domain", "", "servicecert default DNS domain suffix")
	cmd.Flags().StringVar(&roleCertSubjC, "rolecert-subj-c", "", "rolecert default CSR Subject Country")
	cmd.Flags().StringVar(&roleCertSubjP, "rolecert-subj-p", "", "rolecert default CSR Subject Province")
	cmd.Flags().StringVar(&roleCertSubjO, "rolecert-subj-o", "", "rolecert default CSR Subject Organization")
	cmd.Flags().StringVar(&roleCertSubjOU, "rolecert-subj-ou", "", "rolecert default CSR Subject OrganizationalUnit")
	cmd.Flags().BoolVar(&roleCertSpiffe, "rolecert-spiffe", true, "rolecert default: include SPIFFE URI in CSR SAN")
	cmd.Flags().StringVar(&roleCertTrustDomain, "rolecert-spiffe-trust-domain", "", "rolecert default SPIFFE trust domain")
	cmd.Flags().StringVar(&roleCertDNSDomain, "rolecert-dns-domain", "", "rolecert default DNS domain suffix")
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
