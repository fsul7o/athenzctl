package get

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getGroupMeta(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("get group-meta requires NAME")
	}
	g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta, err := groupToMeta(g)
	if err != nil {
		return err
	}
	if done, err := renderStructured(w, format, meta); done || err != nil {
		return err
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"NAME", "SELF-SERVE", "REVIEW-ENABLED", "NOTIFY-ROLES", "USER-AUTHORITY-FILTER"},
		Rows: [][]string{{
			shortName(string(g.Name)),
			boolPtrStr(meta.SelfServe),
			boolPtrStr(meta.ReviewEnabled),
			meta.NotifyRoles,
			meta.UserAuthorityFilter,
		}},
	})
}

func groupToMeta(g *zms.Group) (*zms.GroupMeta, error) {
	b, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	m := &zms.GroupMeta{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}
