package config

import (
	"fmt"

	"github.com/spf13/cobra"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func newDeleteContext(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "delete-context NAME",
		Short: "Delete a context from the athenzctl config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, path, err := loadConfig(opts)
			if err != nil {
				return err
			}
			if !c.Remove(name) {
				return fmt.Errorf("context %q not found", name)
			}
			if err := cfg.Save(path, c); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted context %q\n", name)
			return nil
		},
	}
}
