// Package create implements the `athenzctl create` verb for ZMS resources.
package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `create` command.
func New(opts *cliopts.Options) *cobra.Command {
	kindFlags := cliopts.KindFlagSpec{
		Common: []string{"audit-ref"},
		ByKind: map[resource.Kind][]string{
			resource.KindDomain:         {"admin-users", "description", "parent", "user"},
			resource.KindRole:           {"members", "trust"},
			resource.KindPolicy:         {},
			resource.KindPolicyVersion:  {"from-version"},
			resource.KindService:        {"provider-endpoint", "description"},
			resource.KindServiceKey:     {"pem", "key"},
			resource.KindDomainTemplate: {"param"},
			resource.KindGroup:          {"members"},
			resource.KindMembership:     {"role", "group", "member", "approve", "reject"},
		},
	}
	cmd := &cobra.Command{
		Use:   "create KIND NAME [flags]",
		Short: "Create a new Athenz resource on the server",
		Long: `Create a new Athenz resource. Fails if the resource already exists.

Supported kinds: domain, role, policy, policyversion, service, servicekey,
group, membership, domain-template.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			if err := cliopts.ValidateKindFlags(cmd, "create", kind, kindFlags); err != nil {
				return err
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			auditRef, _ := cmd.Flags().GetString("audit-ref")

			switch kind {
			case resource.KindDomain:
				return createDomain(cmd, out, zc, requireName(args), auditRef)
			case resource.KindRole:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createRole(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindPolicy:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createPolicy(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindPolicyVersion:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createPolicyVersion(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindService:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createService(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindServiceKey:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createServiceKey(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindDomainTemplate:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createDomainTemplate(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindGroup:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createGroup(cmd, out, zc, dom, requireName(args), auditRef)
			case resource.KindMembership:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return createMembership(cmd, out, zc, dom, auditRef)
			default:
				return fmt.Errorf("`create %s` is not supported", kind)
			}
		},
	}

	// Shared flags.
	cmd.Flags().String("audit-ref", "", "audit reference message (required for audit-enabled resources)")

	// Domain flags.
	cmd.Flags().StringSlice("admin-users", nil, "admin users for a new domain (only for KIND=domain)")
	cmd.Flags().String("description", "", "human-readable description")
	cmd.Flags().String("parent", "", "parent domain for a subdomain (only for KIND=domain)")

	// Role / group flags.
	cmd.Flags().StringSlice("members", nil, "comma-separated members (for role/group)")
	cmd.Flags().String("trust", "", "trusted domain for a delegated role (only for KIND=role)")

	// Service flags.
	cmd.Flags().String("provider-endpoint", "", "provider endpoint URL (only for KIND=service)")

	// Membership flags.
	cmd.Flags().String("role", "", "role name (for KIND=membership)")
	cmd.Flags().String("group", "", "group name (for KIND=membership)")
	cmd.Flags().String("member", "", "principal to add (for KIND=membership)")

	// Policyversion flags.
	cmd.Flags().String("from-version", "", "source version to copy from (for KIND=policyversion; defaults to 0)")

	// User domain flag (only for KIND=domain).
	cmd.Flags().String("user", "", "create a user (home.<user>) domain (only for KIND=domain; conflicts with --parent, NAME, --admin-users)")

	// Servicekey flags.
	cmd.Flags().String("pem", "", "path to a PEM-encoded public key file (for KIND=servicekey)")
	cmd.Flags().String("key", "", "inline public key: raw PEM (auto-encoded) or Y-Base64 (for KIND=servicekey)")

	// Domain-template flags.
	cmd.Flags().StringSlice("param", nil, "template parameter KEY=VALUE (repeatable; for KIND=domain-template)")

	// Membership decision flags (approve/reject pending).
	cmd.Flags().Bool("approve", false, "approve a pending membership request (for KIND=membership)")
	cmd.Flags().Bool("reject", false, "reject a pending membership request (for KIND=membership)")

	cliopts.SetKindAwareHelp(cmd, kindFlags)

	return cmd
}

// requireName returns args[1] or "" if not supplied. Kind-specific handlers
// validate whether a positional name is required.
func requireName(args []string) string {
	if len(args) >= 2 {
		return args[1]
	}
	return ""
}
