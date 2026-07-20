package get

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getPolicy(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name != "" {
		p, err := zc.GetPolicy(zms.DomainName(domain), zms.EntityName(name))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, p); done || err != nil {
			return err
		}
		return renderPolicyTable(w, []*zms.Policy{p})
	}
	ps, err := zc.GetPolicies(zms.DomainName(domain), boolPtr(true), nil, "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := printer.WriteStructured(w, format, ps); done || err != nil {
		return err
	}
	return renderPolicyTable(w, ps.List)
}

func renderPolicyTable(w io.Writer, policies []*zms.Policy) error {
	rows := make([][]string, 0, len(policies))
	for _, p := range policies {
		if p == nil {
			continue
		}
		rows = append(rows, []string{
			shortName(string(p.Name)),
			countStr(len(p.Assertions)),
			string(p.Version),
			ts(p.Modified),
		})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "ASSERTIONS", "VERSION", "MODIFIED"},
		Rows:    rows,
	})
}
