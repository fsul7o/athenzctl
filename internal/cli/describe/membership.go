package describe

import (
	"errors"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeMembership(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, member string, format printer.Format) error {
	role, _ := cmd.Flags().GetString("role")
	group, _ := cmd.Flags().GetString("group")
	if role == "" && group == "" {
		return errors.New("describe membership requires --role or --group")
	}
	if role != "" && group != "" {
		return errors.New("--role and --group are mutually exclusive")
	}
	if member == "" {
		return errors.New("describe membership requires NAME (the member principal)")
	}
	if role != "" {
		m, err := zc.GetMembership(zms.DomainName(domain), zms.EntityName(role), zms.MemberName(member), "")
		if err != nil {
			return cliopts.WrapErr(err)
		}
		return render(w, format, m)
	}
	m, err := zc.GetGroupMembership(zms.DomainName(domain), zms.EntityName(group), zms.GroupMemberName(member), "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, m)
}
