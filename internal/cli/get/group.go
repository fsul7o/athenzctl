package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getGroup(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name != "" {
		g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), nil, nil)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := renderStructured(w, format, g); done || err != nil {
			return err
		}
		return renderGroupTable(w, []*zms.Group{g})
	}
	gs, err := zc.GetGroups(zms.DomainName(domain), boolPtr(true), "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, gs); done || err != nil {
		return err
	}
	return renderGroupTable(w, gs.List)
}

func renderGroupTable(w io.Writer, groups []*zms.Group) error {
	rows := make([][]string, 0, len(groups))
	for _, g := range groups {
		if g == nil {
			continue
		}
		rows = append(rows, []string{
			shortName(string(g.Name)),
			countStr(len(g.GroupMembers)),
			ts(g.Modified),
		})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "MEMBERS", "MODIFIED"},
		Rows:    rows,
	})
}
