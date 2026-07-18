package check

import (
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func newAccessCmd(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access ACTION RESOURCE",
		Short: "Check whether a principal is allowed ACTION on RESOURCE",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			action, resource := args[0], args[1]
			principal, _ := cmd.Flags().GetString("principal")
			ext, _ := cmd.Flags().GetBool("ext")

			dom, err := opts.RequireDomain()
			if err != nil {
				return err
			}
			format, err := opts.Format()
			if err != nil {
				return err
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}

			var access *zms.Access
			if ext {
				access, err = zc.GetAccessExt(zms.ActionName(action), resource, zms.DomainName(dom), zms.PrincipalName(principal))
			} else {
				access, err = zc.GetAccess(zms.ActionName(action), zms.ResourceName(resource), zms.DomainName(dom), zms.PrincipalName(principal))
			}
			if err != nil {
				return cliopts.WrapErr(err)
			}
			out := cmd.OutOrStdout()
			switch format {
			case printer.FormatJSON:
				return printer.WriteJSON(out, access)
			case printer.FormatYAML:
				return printer.WriteYAML(out, access)
			}
			verdict := "DENIED"
			if access.Granted {
				verdict = "GRANTED"
			}
			who := principal
			if who == "" {
				who = "(caller)"
			}
			return printer.WriteTable(out, printer.Table{
				Headers: []string{"PRINCIPAL", "ACTION", "RESOURCE", "VERDICT"},
				Rows:    [][]string{{who, action, resource, verdict}},
			})
		},
	}
	cmd.Flags().String("principal", "", "principal to check (default: the caller)")
	cmd.Flags().Bool("ext", false, "use extended access check (allows wildcards / non-standard resource names)")
	return cmd
}
