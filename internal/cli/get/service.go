package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getService(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name != "" {
		s, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(name))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := renderStructured(w, format, s); done || err != nil {
			return err
		}
		return renderServiceTable(w, []*zms.ServiceIdentity{s})
	}
	list, err := zc.GetServiceIdentities(zms.DomainName(domain), boolPtr(false), boolPtr(false), "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, list); done || err != nil {
		return err
	}
	return renderServiceTable(w, list.List)
}

func renderServiceTable(w io.Writer, services []*zms.ServiceIdentity) error {
	rows := make([][]string, 0, len(services))
	for _, s := range services {
		if s == nil {
			continue
		}
		rows = append(rows, []string{
			shortName(string(s.Name)),
			countStr(len(s.PublicKeys)),
			s.ProviderEndpoint,
			ts(s.Modified),
		})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "KEYS", "ENDPOINT", "MODIFIED"},
		Rows:    rows,
	})
}
