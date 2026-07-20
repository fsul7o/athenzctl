package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getTemplate(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	if name != "" {
		tpl, err := zc.GetTemplate(zms.SimpleName(name))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, tpl); done || err != nil {
			return err
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"NAME", "ROLES", "POLICIES", "GROUPS", "SERVICES"},
			Rows: [][]string{{
				name,
				countStr(len(tpl.Roles)),
				countStr(len(tpl.Policies)),
				countStr(len(tpl.Groups)),
				countStr(len(tpl.Services)),
			}},
		})
	}
	list, err := zc.GetServerTemplateList()
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := printer.WriteStructured(w, format, list); done || err != nil {
		return err
	}
	rows := make([][]string, 0, len(list.TemplateNames))
	for _, n := range list.TemplateNames {
		rows = append(rows, []string{string(n)})
	}
	return printer.WriteTable(w, printer.Table{Headers: []string{"NAME"}, Rows: rows})
}
