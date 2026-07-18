package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createService(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	if name == "" {
		return errors.New("create service requires NAME")
	}
	endpoint, _ := cmd.Flags().GetString("provider-endpoint")
	description, _ := cmd.Flags().GetString("description")

	svc := &zms.ServiceIdentity{
		Name:             zms.ServiceName(domain + "." + name),
		Description:      description,
		ProviderEndpoint: endpoint,
	}
	if _, err := zc.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(name), auditRef, cliopts.Ptr(false), "", svc); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "service %q created in domain %s\n", name, domain)
	return nil
}
