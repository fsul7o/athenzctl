package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createMembership(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, auditRef string) error {
	role, _ := cmd.Flags().GetString("role")
	group, _ := cmd.Flags().GetString("group")
	member, _ := cmd.Flags().GetString("member")
	approve, _ := cmd.Flags().GetBool("approve")
	reject, _ := cmd.Flags().GetBool("reject")

	if member == "" {
		return errors.New("create membership requires --member")
	}
	if (role == "") == (group == "") {
		return errors.New("create membership requires exactly one of --role or --group")
	}
	if approve && reject {
		return errors.New("--approve and --reject are mutually exclusive")
	}

	decision := approve || reject
	approved := approve

	if role != "" {
		m := &zms.Membership{
			MemberName: zms.MemberName(member),
			IsMember:   boolPtr(true),
			RoleName:   zms.ResourceName(role),
			Active:     boolPtr(true),
			Approved:   boolPtr(approved),
		}
		if decision {
			if err := zc.PutMembershipDecision(zms.DomainName(domain), zms.EntityName(role), zms.MemberName(member), auditRef, m); err != nil {
				return cliopts.WrapErr(err)
			}
			fmt.Fprintf(w, "%s membership of %s in role %s (domain %s)\n", decisionVerb(approved), member, role, domain)
			return nil
		}
		if _, err := zc.PutMembership(zms.DomainName(domain), zms.EntityName(role), zms.MemberName(member), auditRef, cliopts.Ptr(false), "", m); err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Fprintf(w, "added %s to role %s in domain %s\n", member, role, domain)
		return nil
	}

	gm := &zms.GroupMembership{
		MemberName: zms.GroupMemberName(member),
		IsMember:   boolPtr(true),
		GroupName:  zms.ResourceName(group),
		Active:     boolPtr(true),
		Approved:   boolPtr(approved),
	}
	if decision {
		if err := zc.PutGroupMembershipDecision(zms.DomainName(domain), zms.EntityName(group), zms.GroupMemberName(member), auditRef, gm); err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Fprintf(w, "%s membership of %s in group %s (domain %s)\n", decisionVerb(approved), member, group, domain)
		return nil
	}
	if _, err := zc.PutGroupMembership(zms.DomainName(domain), zms.EntityName(group), zms.GroupMemberName(member), auditRef, cliopts.Ptr(false), "", gm); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "added %s to group %s in domain %s\n", member, group, domain)
	return nil
}

func decisionVerb(approved bool) string {
	if approved {
		return "approved"
	}
	return "rejected"
}

func boolPtr(b bool) *bool { return &b }
