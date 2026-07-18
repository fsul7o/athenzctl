package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCurrentContext(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Display the name of the current context",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := loadConfig(opts)
			if err != nil {
				return err
			}
			if c.CurrentContext == "" {
				return fmt.Errorf("no current context set")
			}
			fmt.Fprintln(cmd.OutOrStdout(), c.CurrentContext)
			return nil
		},
	}
}
