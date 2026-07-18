// Package fetch implements the `athenzctl fetch` verb: sync signed
// artifacts from ZTS to local storage or stdout. ZPU-equivalent behavior.
package fetch

import (
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// New returns the `fetch` command group.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Sync signed artifacts from ZTS to local storage (ZPU equivalent)",
	}
	cmd.AddCommand(newSignedPolicy(opts))
	return cmd
}
