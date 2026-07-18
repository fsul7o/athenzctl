package create

import (
	"errors"
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// createPolicy creates an empty policy. Assertions are added later via
// `edit policy` or, in a future release, `create assertion`.
func createPolicy(_ *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	if name == "" {
		return errors.New("create policy requires NAME")
	}
	policy := &zms.Policy{
		Name:       zms.ResourceName(domain + ":policy." + name),
		Assertions: []*zms.Assertion{},
	}
	if _, err := zc.PutPolicy(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", policy); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "policy %q created in domain %s (empty; use `athenzctl edit policy %s -d %s` to add assertions)\n", name, domain, name, domain)
	return nil
}
