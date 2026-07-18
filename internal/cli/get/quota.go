package get

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func getQuota(w io.Writer, zc *zms.ZMSClient, domain string, format printer.Format) error {
	q, err := zc.GetQuota(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, q); done || err != nil {
		return err
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"DOMAIN", "SUBDOMAIN", "ROLE", "ROLE-MEMBER", "POLICY", "ASSERTION", "SERVICE", "PUBLIC-KEY", "ENTITY", "GROUP", "GROUP-MEMBER"},
		Rows: [][]string{{
			string(q.Name),
			fmt.Sprintf("%d", q.Subdomain),
			fmt.Sprintf("%d", q.Role),
			fmt.Sprintf("%d", q.RoleMember),
			fmt.Sprintf("%d", q.Policy),
			fmt.Sprintf("%d", q.Assertion),
			fmt.Sprintf("%d", q.Service),
			fmt.Sprintf("%d", q.PublicKey),
			fmt.Sprintf("%d", q.Entity),
			fmt.Sprintf("%d", q.Group),
			fmt.Sprintf("%d", q.GroupMember),
		}},
	})
}
