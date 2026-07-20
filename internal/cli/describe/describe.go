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
	kindFlags := cliopts.KindFlagSpec{
		ByKind: map[resource.Kind][]string{
			resource.KindDomain:         {},
			resource.KindDomainMeta:     {},
			resource.KindRole:           {},
			resource.KindRoleMeta:       {},
			resource.KindPolicy:         {},
			resource.KindPolicyVersion:  {},
			resource.KindService:        {},
			resource.KindServiceKey:     {},
			resource.KindGroup:          {},
			resource.KindGroupMeta:      {},
			resource.KindMembership:     {"role", "group"},
			resource.KindTemplate:       {},
			resource.KindDomainTemplate: {},
			resource.KindQuota:          {},
		},
	}
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
			if err := cliopts.ValidateKindFlags(cmd, "describe", kind, kindFlags); err != nil {
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
			domain, err := opts.ResolveResourceDomain(kind, name)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			switch kind {
			case resource.KindDomain:
				return describeDomain(out, zc, name, format)
			case resource.KindDomainMeta:
				return describeDomainMeta(out, zc, domain, format)
			case resource.KindRole:
				return describeRole(out, zc, domain, name, format)
			case resource.KindRoleMeta:
				return describeRoleMeta(out, zc, domain, name, format)
			case resource.KindPolicy:
				return describePolicy(out, zc, domain, name, format)
			case resource.KindPolicyVersion:
				return describePolicyVersion(out, zc, domain, name, format)
			case resource.KindService:
				return describeService(out, zc, domain, name, format)
			case resource.KindServiceKey:
				return describeServiceKey(out, zc, domain, name, format)
			case resource.KindGroup:
				return describeGroup(out, zc, domain, name, format)
			case resource.KindGroupMeta:
				return describeGroupMeta(out, zc, domain, name, format)
			case resource.KindMembership:
				return describeMembership(cmd, out, zc, domain, name, format)
			case resource.KindTemplate:
				return describeTemplate(out, zc, name, format)
			case resource.KindDomainTemplate:
				return describeDomainTemplate(out, zc, domain, name, format)
			case resource.KindQuota:
				return describeQuota(out, zc, domain, format)
			default:
				return fmt.Errorf("`describe %s` is not supported", kind)
			}
		},
	}
	cmd.Flags().String("role", "", "role name to query (only for KIND=membership)")
	cmd.Flags().String("group", "", "group name to query (only for KIND=membership)")
	cliopts.SetKindAwareHelp(cmd, kindFlags)
	return cmd
}

// render dispatches on -o: json/yaml emit the raw object; anything else
// (default, table, wide) goes through the human-readable pretty printer.
func render(w io.Writer, format printer.Format, v any) error {
	if handled, err := printer.WriteStructured(w, format, v); handled || err != nil {
		return err
	}
	return printer.WritePretty(w, v)
}
