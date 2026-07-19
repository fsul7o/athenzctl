// Package cli assembles the athenzctl cobra command tree.
package cli

import (
	"github.com/spf13/cobra"

	checkcmd "github.com/fsul7o/athenzctl/internal/cli/check"
	configcmd "github.com/fsul7o/athenzctl/internal/cli/config"
	createcmd "github.com/fsul7o/athenzctl/internal/cli/create"
	deletecmd "github.com/fsul7o/athenzctl/internal/cli/delete"
	describecmd "github.com/fsul7o/athenzctl/internal/cli/describe"
	editcmd "github.com/fsul7o/athenzctl/internal/cli/edit"
	fetchcmd "github.com/fsul7o/athenzctl/internal/cli/fetch"
	getcmd "github.com/fsul7o/athenzctl/internal/cli/get"
	issuecmd "github.com/fsul7o/athenzctl/internal/cli/issue"
	lookupcmd "github.com/fsul7o/athenzctl/internal/cli/lookup"
	patchcmd "github.com/fsul7o/athenzctl/internal/cli/patch"
	versioncmd "github.com/fsul7o/athenzctl/internal/cli/version"
	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// NewRootCmd builds the top-level `athenzctl` command.
func NewRootCmd() *cobra.Command {
	opts := &cliopts.Options{}
	cmd := &cobra.Command{
		Use:   "athenzctl",
		Short: "Unified command-line client for Athenz",
		Long: `athenzctl is a kubectl-style CLI for the Athenz identity and access
management system. It consolidates the functionality of zms-cli,
zts-accesstoken, zts-svccert, zts-rolecert and zpe-updater behind
a single verb-resource interface.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "path to athenzctl config (default $ATHENZCTL_CONFIG or $HOME/.athenzctl/config.yaml)")
	cmd.PersistentFlags().StringVar(&opts.ContextName, "context", "", "name of the context to use (overrides current-context)")
	cmd.PersistentFlags().StringVarP(&opts.Domain, "domain", "d", "", "Athenz domain scope for the operation")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "output format: table|json|yaml|wide")
	cmd.PersistentFlags().BoolVarP(&opts.InsecureSkipTLSVerify, "insecure-skip-tls-verify", "k", false, "disable TLS certificate and hostname verification")
	cmd.PersistentFlags().StringVarP(&opts.ProxyURL, "proxy", "s", "", "proxy URL (host:port for SOCKS5, or socks5/http/https URL)")
	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		opts.InsecureSkipTLSVerifySet = cmd.Flags().Changed("insecure-skip-tls-verify")
		opts.ProxyURLSet = cmd.Flags().Changed("proxy")
	}

	cmd.AddCommand(versioncmd.New())
	cmd.AddCommand(configcmd.New(&configcmd.Options{ConfigPath: &opts.ConfigPath}))
	cmd.AddCommand(getcmd.New(opts))
	cmd.AddCommand(describecmd.New(opts))
	cmd.AddCommand(createcmd.New(opts))
	cmd.AddCommand(deletecmd.New(opts))
	cmd.AddCommand(editcmd.New(opts))
	cmd.AddCommand(patchcmd.New(opts))
	cmd.AddCommand(checkcmd.New(opts))
	cmd.AddCommand(lookupcmd.New(opts))
	cmd.AddCommand(issuecmd.New(opts))
	cmd.AddCommand(fetchcmd.New(opts))

	return cmd
}
