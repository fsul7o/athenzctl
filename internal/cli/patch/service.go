package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchService(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	orig, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.ServiceIdentity{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if _, err := zc.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(name), auditRef, cliopts.Ptr(false), "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("service %q updated in domain %s\n", name, domain)
	return nil
}
