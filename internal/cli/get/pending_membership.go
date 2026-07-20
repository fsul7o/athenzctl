package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

// getPendingMembership lists pending role or group membership requests. If
// isGroup is false it queries roles; principal is optional and filters the
// view to what a given approver would see (empty = server default).
func getPendingMembership(w io.Writer, zc *zms.ZMSClient, domain string, isGroup bool, principal string, format printer.Format) error {
	if isGroup {
		list, err := zc.GetPendingDomainGroupMembersList(zms.EntityName(principal), domain)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, list); done || err != nil {
			return err
		}
		return renderPendingGroupTable(w, list)
	}
	list, err := zc.GetPendingDomainRoleMembersList(zms.EntityName(principal), domain)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := printer.WriteStructured(w, format, list); done || err != nil {
		return err
	}
	return renderPendingRoleTable(w, list)
}

func renderPendingRoleTable(w io.Writer, list *zms.DomainRoleMembership) error {
	rows := make([][]string, 0)
	for _, drm := range list.DomainRoleMembersList {
		if drm == nil {
			continue
		}
		for _, m := range drm.Members {
			if m == nil {
				continue
			}
			for _, r := range m.MemberRoles {
				if r == nil {
					continue
				}
				rows = append(rows, []string{
					string(drm.DomainName),
					string(r.RoleName),
					string(m.MemberName),
					ts(r.Expiration),
					r.AuditRef,
				})
			}
		}
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"DOMAIN", "ROLE", "MEMBER", "EXPIRATION", "AUDIT-REF"},
		Rows:    rows,
	})
}

func renderPendingGroupTable(w io.Writer, list *zms.DomainGroupMembership) error {
	rows := make([][]string, 0)
	for _, dgm := range list.DomainGroupMembersList {
		if dgm == nil {
			continue
		}
		for _, m := range dgm.Members {
			if m == nil {
				continue
			}
			for _, g := range m.MemberGroups {
				if g == nil {
					continue
				}
				rows = append(rows, []string{
					string(dgm.DomainName),
					string(g.GroupName),
					string(m.MemberName),
					ts(g.Expiration),
					g.AuditRef,
				})
			}
		}
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"DOMAIN", "GROUP", "MEMBER", "EXPIRATION", "AUDIT-REF"},
		Rows:    rows,
	})
}
