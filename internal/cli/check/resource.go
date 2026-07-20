package check

import (
	"errors"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func newResourceCmd(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "List resources a principal can access",
		RunE: func(cmd *cobra.Command, args []string) error {
			principal, _ := cmd.Flags().GetString("principal")
			action, _ := cmd.Flags().GetString("action")
			filter, _ := cmd.Flags().GetString("role-filter")
			if principal == "" {
				return errors.New("check resource requires --principal")
			}

			format, err := opts.Format()
			if err != nil {
				return err
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			list, err := zc.GetResourceAccessList(zms.PrincipalName(principal), zms.ActionName(action), filter)
			if err != nil {
				return cliopts.WrapErr(err)
			}
			out := cmd.OutOrStdout()
			if handled, err := printer.WriteStructured(out, format, list); handled || err != nil {
				return err
			}
			rows := make([][]string, 0)
			for _, r := range list.Resources {
				if r == nil {
					continue
				}
				for _, a := range r.Assertions {
					effect := "ALLOW"
					if a.Effect != nil {
						effect = a.Effect.String()
					}
					rows = append(rows, []string{string(r.Principal), effect, string(a.Action), string(a.Resource), string(a.Role)})
				}
			}
			return printer.WriteTable(out, printer.Table{
				Headers: []string{"PRINCIPAL", "EFFECT", "ACTION", "RESOURCE", "ROLE"},
				Rows:    rows,
			})
		},
	}
	cmd.Flags().String("principal", "", "principal to enumerate access for (required)")
	cmd.Flags().String("action", "", "restrict to a single action (optional)")
	cmd.Flags().String("role-filter", "", "server-side role name filter (optional)")
	return cmd
}
