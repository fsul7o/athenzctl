package config

import (
	"fmt"

	"github.com/spf13/cobra"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

func newUseContext(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "use-context NAME",
		Short: "Set the current-context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			c, path, err := loadConfig(opts)
			if err != nil {
				return err
			}
			if c.Find(name) == nil {
				return fmt.Errorf("context %q not found", name)
			}
			c.CurrentContext = name
			if err := cfg.Save(path, c); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q\n", name)
			return nil
		},
	}
}
