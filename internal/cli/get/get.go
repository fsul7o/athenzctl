// Package get implements the `athenzctl get` verb for ZMS resources.
package get

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `get` cobra command.
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
			resource.KindTemplate:       {},
			resource.KindDomainTemplate: {},
			resource.KindQuota:          {},
			resource.KindGroup:          {},
			resource.KindGroupMeta:      {},
			resource.KindMembership:     {"role", "group", "pending", "principal"},
		},
	}
	cmd := &cobra.Command{
		Use:   "get KIND [NAME]",
		Short: "Display one or many Athenz resources",
		Long: `Display one or many Athenz resources.

Supported kinds: domain(s), domain-meta, role(s), role-meta, policy(ies),
policyversion(s), service(s), servicekey(s), group(s), group-meta,
membership(s), template(s), domain-template(s), quota.`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			if err := cliopts.ValidateKindFlags(cmd, "get", kind, kindFlags); err != nil {
				return err
			}
			format, err := opts.Format()
			if err != nil {
				return err
			}
			var name string
			if len(args) == 2 {
				name = args[1]
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			switch kind {
			case resource.KindDomain:
				return getDomain(out, zc, name, format)
			case resource.KindDomainMeta:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return getDomainMeta(out, zc, dom, format)
			case resource.KindRole:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getRole(out, zc, dom, name, format)
			case resource.KindRoleMeta:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getRoleMeta(out, zc, dom, name, format)
			case resource.KindPolicy:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getPolicy(out, zc, dom, name, format)
			case resource.KindPolicyVersion:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getPolicyVersion(out, zc, dom, name, format)
			case resource.KindService:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getService(out, zc, dom, name, format)
			case resource.KindServiceKey:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getServiceKey(out, zc, dom, name, format)
			case resource.KindTemplate:
				return getTemplate(out, zc, name, format)
			case resource.KindDomainTemplate:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getDomainTemplate(out, zc, dom, format)
			case resource.KindQuota:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return getQuota(out, zc, dom, format)
			case resource.KindGroup:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getGroup(out, zc, dom, name, format)
			case resource.KindGroupMeta:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getGroupMeta(out, zc, dom, name, format)
			case resource.KindMembership:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return getMembership(cmd, out, zc, dom, name, format)
			default:
				return fmt.Errorf("`get %s` is not supported", kind)
			}
		},
	}
	cmd.Flags().String("role", "", "role name to query (only for KIND=membership)")
	cmd.Flags().String("group", "", "group name to query (only for KIND=membership)")
	cmd.Flags().Bool("pending", false, "list pending membership requests in the current domain (only for KIND=membership)")
	cmd.Flags().String("principal", "", "filter pending membership by approving principal (only with --pending)")
	cliopts.SetKindAwareHelp(cmd, kindFlags)
	return cmd
}

// renderStructured picks JSON or YAML for -o flags that request them. Returns
// (rendered, err): rendered=true means the caller should not fall through to
// a tabular renderer.
func renderStructured(w io.Writer, format printer.Format, v any) (bool, error) {
	switch format {
	case printer.FormatJSON:
		return true, printer.WriteJSON(w, v)
	case printer.FormatYAML:
		return true, printer.WriteYAML(w, v)
	}
	return false, nil
}
