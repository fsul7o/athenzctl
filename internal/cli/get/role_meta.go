package get

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getRoleMeta(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("get role-meta requires NAME")
	}
	r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), nil, nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta, err := roleToMeta(r)
	if err != nil {
		return err
	}
	if done, err := renderStructured(w, format, meta); done || err != nil {
		return err
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "SELF-SERVE", "REVIEW-ENABLED", "MEMBER-EXPIRY-DAYS", "NOTIFY-ROLES"},
		Rows: [][]string{{
			shortName(string(r.Name)),
			boolPtrStr(meta.SelfServe),
			boolPtrStr(meta.ReviewEnabled),
			i32Str(meta.MemberExpiryDays),
			meta.NotifyRoles,
		}},
	})
}

func roleToMeta(r *zms.Role) (*zms.RoleMeta, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	m := &zms.RoleMeta{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}

func boolPtrStr(p *bool) string {
	if p == nil {
		return "-"
	}
	if *p {
		return "true"
	}
	return "false"
}
