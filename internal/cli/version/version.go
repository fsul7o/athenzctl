package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/version"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print athenzctl version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "athenzctl %s (commit %s, built %s)\n", version.Version, version.Commit, version.Date)
			return err
		},
	}
}
