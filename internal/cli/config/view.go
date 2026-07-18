package config

import (
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newView(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display the merged athenzctl config",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := loadConfig(opts)
			if err != nil {
				return err
			}
			enc := yaml.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent(2)
			defer enc.Close()
			return enc.Encode(c)
		},
	}
}
