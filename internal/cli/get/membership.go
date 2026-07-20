package get

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getMembership(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, member string, format printer.Format) error {
	role, _ := cmd.Flags().GetString("role")
	group, _ := cmd.Flags().GetString("group")
	pending, _ := cmd.Flags().GetBool("pending")
	principal, _ := cmd.Flags().GetString("principal")

	if pending {
		return getPendingMembership(w, zc, domain, group != "", principal, format)
	}

	if role == "" && group == "" {
		return errors.New("get membership requires --role, --group, or --pending")
	}
	if role != "" && group != "" {
		return errors.New("--role and --group are mutually exclusive")
	}

	switch {
	case role != "" && member != "":
		m, err := zc.GetMembership(zms.DomainName(domain), zms.EntityName(role), zms.MemberName(member), "")
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, m); done || err != nil {
			return err
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"ROLE", "MEMBER", "IS-MEMBER", "EXPIRATION"},
			Rows: [][]string{{
				role, member, fmt.Sprintf("%v", isTrue(m.IsMember)), ts(m.Expiration),
			}},
		})
	case group != "" && member != "":
		m, err := zc.GetGroupMembership(zms.DomainName(domain), zms.EntityName(group), zms.GroupMemberName(member), "")
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, m); done || err != nil {
			return err
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"GROUP", "MEMBER", "IS-MEMBER", "EXPIRATION"},
			Rows: [][]string{{
				group, member, fmt.Sprintf("%v", isTrue(m.IsMember)), ts(m.Expiration),
			}},
		})
	case role != "":
		r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(role), nil, nil, nil)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, r); done || err != nil {
			return err
		}
		rows := make([][]string, 0, len(r.RoleMembers))
		for _, m := range r.RoleMembers {
			rows = append(rows, []string{string(m.MemberName), ts(m.Expiration), memberStatus(m)})
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"MEMBER", "EXPIRATION", "STATUS"},
			Rows:    rows,
		})
	default: // group != ""
		g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(group), nil, nil)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, g); done || err != nil {
			return err
		}
		rows := make([][]string, 0, len(g.GroupMembers))
		for _, m := range g.GroupMembers {
			rows = append(rows, []string{string(m.MemberName), ts(m.Expiration), groupMemberStatus(m)})
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"MEMBER", "EXPIRATION", "STATUS"},
			Rows:    rows,
		})
	}
}

func isTrue(b *bool) bool { return b != nil && *b }

func memberStatus(m *zms.RoleMember) string {
	if m == nil {
		return ""
	}
	if m.Approved != nil && !*m.Approved {
		return "pending"
	}
	return "approved"
}

func groupMemberStatus(m *zms.GroupMember) string {
	if m == nil {
		return ""
	}
	if m.Approved != nil && !*m.Approved {
		return "pending"
	}
	return "approved"
}
