package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createDomain(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, name, auditRef string) error {
	user, _ := cmd.Flags().GetString("user")
	admins, _ := cmd.Flags().GetStringSlice("admin-users")
	desc, _ := cmd.Flags().GetString("description")
	parent, _ := cmd.Flags().GetString("parent")

	if user != "" {
		if name != "" || parent != "" || len(admins) > 0 {
			return errors.New("create domain --user takes no NAME, --parent, or --admin-users")
		}
		detail := &zms.UserDomain{
			Name:        zms.SimpleName(user),
			Description: desc,
		}
		created, err := zc.PostUserDomain(zms.SimpleName(user), auditRef, "", detail)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Fprintf(w, "user domain %q created\n", string(created.Name))
		return nil
	}

	if name == "" {
		return errors.New("create domain requires NAME (or --user)")
	}
	if len(admins) == 0 {
		return errors.New("create domain requires --admin-users")
	}

	adminNames := make([]zms.ResourceName, 0, len(admins))
	for _, a := range admins {
		adminNames = append(adminNames, zms.ResourceName(a))
	}

	if parent == "" {
		detail := &zms.TopLevelDomain{
			Name:        zms.SimpleName(name),
			AdminUsers:  adminNames,
			Description: desc,
		}
		created, err := zc.PostTopLevelDomain(auditRef, "", detail)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Fprintf(w, "domain %q created\n", string(created.Name))
		return nil
	}

	detail := &zms.SubDomain{
		Name:        zms.SimpleName(name),
		Parent:      zms.DomainName(parent),
		AdminUsers:  adminNames,
		Description: desc,
	}
	created, err := zc.PostSubDomain(zms.DomainName(parent), auditRef, "", detail)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "subdomain %q created under %s\n", string(created.Name), parent)
	return nil
}
