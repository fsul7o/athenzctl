// Package issue implements the `athenzctl issue` verb: ask ZTS to mint a
// short-lived credential (access token, role cert) for the caller, or fetch
// an instance register token (bootstrap attestation used with auth-mode
// "copperargos"). Not a resource operation — the result is written to
// stdout or a file rather than persisted on the server.
//
// Service X.509 identity certificates are intentionally out of scope here;
// use the official zts-svccert tool for one-off/manual service cert
// issuance, or configure auth-mode "ntoken"/"copperargos" (see
// internal/ntokenauth, internal/copperargosauth) to mint them automatically
// on every athenzctl invocation.
package issue

import (
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// New returns the `issue` command group.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Ask ZTS to issue a credential (access token, role cert) or an instance register token",
	}
	cmd.AddCommand(newAccessToken(opts))
	cmd.AddCommand(newRoleCert(opts))
	cmd.AddCommand(newInstanceRegisterToken(opts))
	return cmd
}
