// Package delete implements the `athenzctl delete` verb for ZMS resources.
package delete

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `delete` command.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete KIND [NAME] [flags]",
		Short: "Delete an Athenz resource",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			auditRef, _ := cmd.Flags().GetString("audit-ref")
			name := ""
			if len(args) >= 2 {
				name = args[1]
			}

			switch kind {
			case resource.KindDomain:
				user, _ := cmd.Flags().GetString("user")
				parent, _ := cmd.Flags().GetString("parent")
				if user != "" {
					if parent != "" || name != "" {
						return errors.New("delete domain --user takes no NAME or --parent")
					}
					if err := zc.DeleteUserDomain(zms.SimpleName(user), auditRef, ""); err != nil {
						return cliopts.WrapErr(err)
					}
					fmt.Fprintf(out, "user domain home.%s deleted\n", user)
					return nil
				}
				if name == "" {
					return errors.New("delete domain requires NAME (or --user)")
				}
				if parent == "" {
					if err := zc.DeleteTopLevelDomain(zms.SimpleName(name), auditRef, ""); err != nil {
						return cliopts.WrapErr(err)
					}
					fmt.Fprintf(out, "domain %q deleted\n", name)
				} else {
					if err := zc.DeleteSubDomain(zms.DomainName(parent), zms.SimpleName(name), auditRef, ""); err != nil {
						return cliopts.WrapErr(err)
					}
					fmt.Fprintf(out, "subdomain %q deleted from %s\n", name, parent)
				}
				return nil

			case resource.KindRole:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete role requires NAME")
				}
				if err := zc.DeleteRole(zms.DomainName(dom), zms.EntityName(name), auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "role %q deleted from domain %s\n", name, dom)
				return nil

			case resource.KindPolicy:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete policy requires NAME")
				}
				if err := zc.DeletePolicy(zms.DomainName(dom), zms.EntityName(name), auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "policy %q deleted from domain %s\n", name, dom)
				return nil

			case resource.KindService:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete service requires NAME")
				}
				if err := zc.DeleteServiceIdentity(zms.DomainName(dom), zms.SimpleName(name), auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "service %q deleted from domain %s\n", name, dom)
				return nil

			case resource.KindGroup:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete group requires NAME")
				}
				if err := zc.DeleteGroup(zms.DomainName(dom), zms.EntityName(name), auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "group %q deleted from domain %s\n", name, dom)
				return nil

			case resource.KindMembership:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				return deleteMembership(cmd, out, zc, dom, auditRef)

			case resource.KindServiceKey:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete servicekey requires SERVICE:KEYID")
				}
				ref, err := resource.ParseServiceKey(name)
				if err != nil {
					return err
				}
				if ref.KeyID == "" {
					return errors.New("delete servicekey requires SERVICE:KEYID (missing :KEYID)")
				}
				if err := zc.DeletePublicKeyEntry(zms.DomainName(dom), zms.SimpleName(ref.Service), ref.KeyID, auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "servicekey %s:%s deleted from domain %s\n", ref.Service, ref.KeyID, dom)
				return nil

			case resource.KindDomainTemplate:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete domain-template requires template NAME")
				}
				if err := zc.DeleteDomainTemplate(zms.DomainName(dom), zms.SimpleName(name), auditRef); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "domain-template %q removed from domain %s\n", name, dom)
				return nil

			case resource.KindQuota:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if err := zc.DeleteQuota(zms.DomainName(dom), auditRef); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "quota reset to defaults for domain %s\n", dom)
				return nil

			case resource.KindPolicyVersion:
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				if name == "" {
					return errors.New("delete policyversion requires POLICY:VERSION")
				}
				ref, err := resource.ParsePolicyVersion(name)
				if err != nil {
					return err
				}
				if ref.Version == "" {
					return errors.New("delete policyversion requires POLICY:VERSION (missing :VERSION)")
				}
				if err := zc.DeletePolicyVersion(zms.DomainName(dom), zms.EntityName(ref.Policy), zms.SimpleName(ref.Version), auditRef, ""); err != nil {
					return cliopts.WrapErr(err)
				}
				fmt.Fprintf(out, "policy version %s:%s deleted from domain %s\n", ref.Policy, ref.Version, dom)
				return nil

			default:
				return fmt.Errorf("`delete %s` is not supported", kind)
			}
		},
	}

	cmd.Flags().String("audit-ref", "", "audit reference message")
	cmd.Flags().String("parent", "", "parent domain (only for KIND=domain, when deleting a subdomain)")
	cmd.Flags().String("user", "", "user (home.<user>) domain to delete (only for KIND=domain; conflicts with NAME and --parent)")
	cmd.Flags().String("role", "", "role name (for KIND=membership)")
	cmd.Flags().String("group", "", "group name (for KIND=membership)")
	cmd.Flags().String("member", "", "principal to remove (for KIND=membership)")
	return cmd
}

func deleteMembership(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, auditRef string) error {
	role, _ := cmd.Flags().GetString("role")
	group, _ := cmd.Flags().GetString("group")
	member, _ := cmd.Flags().GetString("member")
	if member == "" {
		return errors.New("delete membership requires --member")
	}
	if (role == "") == (group == "") {
		return errors.New("delete membership requires exactly one of --role or --group")
	}

	if role != "" {
		if err := zc.DeleteMembership(zms.DomainName(domain), zms.EntityName(role), zms.MemberName(member), auditRef, ""); err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Fprintf(w, "removed %s from role %s in domain %s\n", member, role, domain)
		return nil
	}
	if err := zc.DeleteGroupMembership(zms.DomainName(domain), zms.EntityName(group), zms.GroupMemberName(member), auditRef, ""); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "removed %s from group %s in domain %s\n", member, group, domain)
	return nil
}
