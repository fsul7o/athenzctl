// Package issue implements the `athenzctl issue` verb: ask ZTS to mint a
// short-lived credential (access token, service cert, role cert) for the
// caller. Not a resource operation — the result is written to stdout or a
// file rather than persisted on the server.
package issue

import (
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// New returns the `issue` command group.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Ask ZTS to issue a credential (access token, service cert, role cert)",
	}
	cmd.AddCommand(newAccessToken(opts))
	cmd.AddCommand(newRoleCert(opts))
	cmd.AddCommand(newServiceCert(opts))
	cmd.AddCommand(newInstanceRegisterToken(opts))
	return cmd
}
