package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchRole(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	trueVal := true
	orig, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), &trueVal, nil, &trueVal)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.Role{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if _, err := zc.PutRole(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("role %q updated in domain %s\n", name, domain)
	return nil
}
