// Package describe implements the `athenzctl describe` verb. By default
// it renders every field returned by ZMS in a human-readable indented
// tree; `-o yaml` and `-o json` emit the raw response in those formats.
// The pretty renderer discovers fields via JSON marshaling, so new
// upstream fields surface automatically.
package describe

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `describe` command.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe KIND [NAME]",
		Short: "Show detailed information about a single Athenz resource",
		Long: `Show detailed information about an Athenz resource.

Supported kinds: domain, domain-meta, role, role-meta, policy,
policyversion, service, servicekey, group, group-meta, membership,
template, domain-template, quota.

NAME is required for most kinds. For quota (one per domain) NAME is
ignored. For membership, use --role or --group with NAME as the member
principal.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			name := ""
			if len(args) == 2 {
				name = args[1]
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			format, err := opts.Format()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			switch kind {
			case resource.KindDomain:
				return describeDomain(out, zc, name, format)
			case resource.KindDomainMeta:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return describeDomainMeta(out, zc, dom, format)
			case resource.KindRole:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeRole(out, zc, dom, name, format)
			case resource.KindRoleMeta:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeRoleMeta(out, zc, dom, name, format)
			case resource.KindPolicy:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describePolicy(out, zc, dom, name, format)
			case resource.KindPolicyVersion:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describePolicyVersion(out, zc, dom, name, format)
			case resource.KindService:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeService(out, zc, dom, name, format)
			case resource.KindServiceKey:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeServiceKey(out, zc, dom, name, format)
			case resource.KindGroup:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeGroup(out, zc, dom, name, format)
			case resource.KindGroupMeta:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeGroupMeta(out, zc, dom, name, format)
			case resource.KindMembership:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeMembership(cmd, out, zc, dom, name, format)
			case resource.KindTemplate:
				return describeTemplate(out, zc, name, format)
			case resource.KindDomainTemplate:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return describeDomainTemplate(out, zc, dom, name, format)
			case resource.KindQuota:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return describeQuota(out, zc, dom, format)
			default:
				return fmt.Errorf("`describe %s` is not supported", kind)
			}
		},
	}
	cmd.Flags().String("role", "", "role name to query (only for KIND=membership)")
	cmd.Flags().String("group", "", "group name to query (only for KIND=membership)")
	return cmd
}

// render dispatches on -o: json/yaml emit the raw object; anything else
// (default, table, wide) goes through the human-readable pretty printer.
func render(w io.Writer, format printer.Format, v any) error {
	switch format {
	case printer.FormatJSON:
		return printer.WriteJSON(w, v)
	case printer.FormatYAML:
		return printer.WriteYAML(w, v)
	}
	return printer.WritePretty(w, v)
}
