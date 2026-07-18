package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchPolicy(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	orig, err := zc.GetPolicy(zms.DomainName(domain), zms.EntityName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.Policy{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if _, err := zc.PutPolicy(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("policy %q updated in domain %s\n", name, domain)
	return nil
}
