package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getDomainTemplate(w io.Writer, zc *zms.ZMSClient, domain string, format printer.Format) error {
	list, err := zc.GetDomainTemplateList(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, list); done || err != nil {
		return err
	}
	rows := make([][]string, 0, len(list.TemplateNames))
	for _, n := range list.TemplateNames {
		rows = append(rows, []string{string(n)})
	}
	return printer.WriteTable(w, printer.Table{Headers: []string{"NAME"}, Rows: rows})
}
