package issue

import (
	"errors"
	"fmt"
	"os"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// newInstanceRegisterToken fetches an instance register token from ZTS.
// The token is a one-time bootstrap credential valid for 30 minutes that is
// used as attestation data for a subsequent Copper Argos provider
// registration call (see zts-svccert -get-instance-register-token, or
// athenzctl's own `config set-context --auth-mode copperargos
// --copperargos-auth-attestation-data`, which consumes a token saved via
// --out here).
func newInstanceRegisterToken(opts *cliopts.Options) *cobra.Command {
	var (
		service  string
		provider string
		instance string
		outPath  string
	)
	cmd := &cobra.Command{
		Use:   "instance-register-token",
		Short: "Fetch an instance register token (bootstrap attestation) from ZTS",
		RunE: func(cmd *cobra.Command, _ []string) error {
			domain, err := opts.RequireDomain()
			if err != nil {
				return err
			}
			if service == "" {
				return errors.New("issue instance-register-token requires --service")
			}
			if provider == "" {
				return errors.New("issue instance-register-token requires --provider")
			}
			if instance == "" {
				return errors.New("issue instance-register-token requires --instance")
			}
			zc, err := opts.ZTSClient()
			if err != nil {
				return err
			}
			tok, err := zc.GetInstanceRegisterToken(
				zts.ServiceName(provider),
				zts.DomainName(domain),
				zts.SimpleName(service),
				zts.PathElement(instance),
			)
			if err != nil {
				return cliopts.WrapErr(err)
			}
			if outPath != "" && outPath != "-" {
				return os.WriteFile(outPath, []byte(tok.AttestationData), 0o600)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), tok.AttestationData)
			return err
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "Athenz provider service name (required)")
	cmd.Flags().StringVar(&instance, "instance", "", "instance ID (required)")
	cmd.Flags().StringVar(&outPath, "out", "", "path to write the token (default: stdout)")
	return cmd
}
