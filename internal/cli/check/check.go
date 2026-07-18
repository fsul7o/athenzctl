// Package check implements the `athenzctl check` verb: authorization
// queries (access checks, resource-access listing). The verb is a
// standalone group because these are operations, not resources.
package check

import (
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// New returns the `check` command tree.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Authorization checks against Athenz",
		Long: `Query Athenz authorization state.

Subcommands:
  access    - check whether a principal is allowed ACTION on RESOURCE
  resource  - list all resources a principal has access to`,
	}
	cmd.AddCommand(newAccessCmd(opts))
	cmd.AddCommand(newResourceCmd(opts))
	return cmd
}
