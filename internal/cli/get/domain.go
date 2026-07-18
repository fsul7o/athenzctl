package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getDomain(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	if name != "" {
		d, err := zc.GetDomain(zms.DomainName(name))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := renderStructured(w, format, d); done || err != nil {
			return err
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"NAME", "MODIFIED", "DESCRIPTION"},
			Rows:    [][]string{{string(d.Name), ts(d.Modified), d.Description}},
		})
	}
	list, err := zc.GetDomainList(nil, "", "", nil, "", nil, "", "", "", "", "", "", "", "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, list); done || err != nil {
		return err
	}
	rows := make([][]string, 0, len(list.Names))
	for _, n := range list.Names {
		rows = append(rows, []string{string(n)})
	}
	return printer.WriteTable(w, printer.Table{Headers: []string{"NAME"}, Rows: rows})
}
