// Package lookup implements the `athenzctl lookup` verb.
package lookup

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `lookup` cobra command. The command accepts a resource kind
// as its argument so additional lookupable kinds can be added later without
// changing the top-level command shape.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup KIND",
		Short: "Find Athenz resources by supported server-side fields",
		Long: `Find Athenz resources using a server-side field selector.

Currently supported kinds: domain.

For domain lookups, --field-selector accepts comma-separated equality
expressions using the fields supported by ZMS, for example:

  athenzctl lookup domain --field-selector member=user.jane,role=admin
  athenzctl lookup domain --field-selector tagKey=team,tagValue=security
  athenzctl lookup domain --field-selector account=1234567890`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			if kind != resource.KindDomain {
				return fmt.Errorf("`lookup %s` is not supported", kind)
			}

			fieldSelector, err := cmd.Flags().GetString("field-selector")
			if err != nil {
				return err
			}
			selector, err := parseDomainSelector(fieldSelector)
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
			return lookupDomain(cmd.OutOrStdout(), zc, selector, format)
		},
	}
	cmd.Flags().String("field-selector", "", "comma-separated ZMS field selector (required for KIND=domain)")
	return cmd
}

// lookupDomain performs the one ZMS domain-list request for a validated
// selector and renders the result using the common athenzctl output formats.
func lookupDomain(w io.Writer, zc *zms.ZMSClient, selector domainSelector, format printer.Format) error {
	list, err := zc.GetDomainList(
		nil,
		"",
		"",
		nil,
		selector.account,
		selector.productNumber,
		selector.roleMember,
		selector.roleName,
		selector.azure,
		selector.gcp,
		selector.tagKey,
		selector.tagValue,
		selector.businessService,
		selector.productID,
		"",
	)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if handled, err := printer.WriteStructured(w, format, list); handled || err != nil {
		return err
	}

	rows := make([][]string, 0, len(list.Names))
	for _, name := range list.Names {
		rows = append(rows, []string{string(name)})
	}
	return printer.WriteTable(w, printer.Table{Headers: []string{"NAME"}, Rows: rows})
}
