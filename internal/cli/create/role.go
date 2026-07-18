package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createRole(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	if name == "" {
		return errors.New("create role requires NAME")
	}
	members, _ := cmd.Flags().GetStringSlice("members")
	trust, _ := cmd.Flags().GetString("trust")

	role := &zms.Role{
		Name:  zms.ResourceName(domain + ":role." + name),
		Trust: zms.DomainName(trust),
	}
	if trust == "" {
		roleMembers := make([]*zms.RoleMember, 0, len(members))
		for _, m := range members {
			roleMembers = append(roleMembers, &zms.RoleMember{MemberName: zms.MemberName(m)})
		}
		role.RoleMembers = roleMembers
	}

	if _, err := zc.PutRole(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", role); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "role %q created in domain %s\n", name, domain)
	return nil
}
