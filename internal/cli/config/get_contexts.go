package config

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newGetContexts(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get-contexts",
		Short: "List all contexts defined in the athenzctl config",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := loadConfig(opts)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "CURRENT\tNAME\tZMS\tZTS")
			for _, ctx := range c.Contexts {
				marker := " "
				if ctx.Name == c.CurrentContext {
					marker = "*"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, ctx.Name, ctx.ZMSURL, ctx.ZTSURL)
			}
			return w.Flush()
		},
	}
}
