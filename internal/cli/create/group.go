package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createGroup(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	if name == "" {
		return errors.New("create group requires NAME")
	}
	members, _ := cmd.Flags().GetStringSlice("members")

	groupMembers := make([]*zms.GroupMember, 0, len(members))
	for _, m := range members {
		groupMembers = append(groupMembers, &zms.GroupMember{MemberName: zms.GroupMemberName(m)})
	}
	g := &zms.Group{
		Name:         zms.ResourceName(domain + ":group." + name),
		GroupMembers: groupMembers,
	}
	if _, err := zc.PutGroup(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", g); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "group %q created in domain %s\n", name, domain)
	return nil
}
