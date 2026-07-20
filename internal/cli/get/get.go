// Package get implements the `athenzctl get` verb for ZMS resources.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
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
			domain, err := opts.ResolveResourceDomain(kind, name)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			switch kind {
			case resource.KindDomain:
				return getDomain(out, zc, name, format)
			case resource.KindDomainMeta:
				return getDomainMeta(out, zc, domain, format)
			case resource.KindRole:
				return getRole(out, zc, domain, name, format)
			case resource.KindRoleMeta:
				return getRoleMeta(out, zc, domain, name, format)
			case resource.KindPolicy:
				return getPolicy(out, zc, domain, name, format)
			case resource.KindPolicyVersion:
				return getPolicyVersion(out, zc, domain, name, format)
			case resource.KindService:
				return getService(out, zc, domain, name, format)
			case resource.KindServiceKey:
				return getServiceKey(out, zc, domain, name, format)
			case resource.KindTemplate:
				return getTemplate(out, zc, name, format)
			case resource.KindDomainTemplate:
				return getDomainTemplate(out, zc, domain, format)
			case resource.KindQuota:
				return getQuota(out, zc, domain, format)
			case resource.KindGroup:
				return getGroup(out, zc, domain, name, format)
			case resource.KindGroupMeta:
				return getGroupMeta(out, zc, domain, name, format)
			case resource.KindMembership:
				return getMembership(cmd, out, zc, domain, name, format)
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
