package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getRole(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name != "" {
		r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), nil, nil, nil)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := renderStructured(w, format, r); done || err != nil {
			return err
		}
		return renderRoleTable(w, []*zms.Role{r})
	}
	rs, err := zc.GetRoles(zms.DomainName(domain), boolPtr(true), "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, rs); done || err != nil {
		return err
	}
	return renderRoleTable(w, rs.List)
}

func renderRoleTable(w io.Writer, roles []*zms.Role) error {
	rows := make([][]string, 0, len(roles))
	for _, r := range roles {
		if r == nil {
			continue
		}
		members := len(r.RoleMembers)
		if members == 0 {
			members = len(r.Members)
		}
		rows = append(rows, []string{
			shortName(string(r.Name)),
			countStr(members),
			string(r.Trust),
			ts(r.Modified),
		})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "MEMBERS", "TRUST", "MODIFIED"},
		Rows:    rows,
	})
}

func boolPtr(b bool) *bool { return &b }
