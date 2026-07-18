package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchRoleMeta(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), nil, nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	orig := &zms.RoleMeta{}
	if err := jsonRoundTrip(r, orig); err != nil {
		return err
	}
	out := &zms.RoleMeta{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if err := zc.PutRoleMeta(zms.DomainName(domain), zms.EntityName(name), auditRef, "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("role-meta %q updated in domain %s\n", name, domain)
	return nil
}
