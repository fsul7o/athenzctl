package get

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getDomainMeta(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	d, err := zc.GetDomain(zms.DomainName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta, err := domainToMeta(d)
	if err != nil {
		return err
	}
	if done, err := renderStructured(w, format, meta); done || err != nil {
		return err
	}
	return renderDomainMetaTable(w, name, meta)
}

// domainToMeta extracts DomainMeta from a Domain via JSON round-trip. Domain
// and DomainMeta share JSON tags for every meta field, so this is exact.
func domainToMeta(d *zms.Domain) (*zms.DomainMeta, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	m := &zms.DomainMeta{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}

func renderDomainMetaTable(w io.Writer, name string, m *zms.DomainMeta) error {
	rows := [][]string{
		{name, m.Description, string(m.Org), i32Str(m.MemberExpiryDays), i32Str(m.TokenExpiryMins), m.BusinessService},
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "DESCRIPTION", "ORG", "MEMBER-EXPIRY-DAYS", "TOKEN-EXPIRY-MINS", "BUSINESS-SERVICE"},
		Rows:    rows,
	})
}

func i32Str(p *int32) string {
	if p == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *p)
}
