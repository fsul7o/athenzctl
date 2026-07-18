package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func newSetContext(opts *Options) *cobra.Command {
	var (
		zmsURL        string
		ztsURL        string
		cert          string
		key           string
		caCert        string
		zmsServerName string
		ztsServerName string
		authMode      string
		// exec fields
		execCommand  string
		execArgs     []string
		execEnv      []string
		execCertPath string
		execKeyPath  string
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
			if cmd.Flags().Changed("auth-mode") {
				ctx.AuthMode = authMode
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
	cmd.Flags().StringVar(&authMode, "auth-mode", "", "authentication mode: \"\" or \"mtls\" (default), \"exec\" (obtain the client cert by execing an external command that places it at a known path)")
	// exec flags
	cmd.Flags().StringVar(&execCommand, "exec-command", "", "exec: path to the external command that places a fresh cert/key at exec-cert-path/exec-key-path")
	cmd.Flags().StringArrayVar(&execArgs, "exec-arg", nil, "exec: argument to pass to the exec command (repeatable, replaces the full list when set)")
	cmd.Flags().StringArrayVar(&execEnv, "exec-env", nil, "exec: KEY=VALUE environment variable to set for the exec command (repeatable, replaces the full map when set)")
	cmd.Flags().StringVar(&execCertPath, "exec-cert-path", "", "exec: path the exec command writes the cert PEM to, read back after it exits")
	cmd.Flags().StringVar(&execKeyPath, "exec-key-path", "", "exec: path the exec command writes the key PEM to, read back after it exits")
	return cmd
}
