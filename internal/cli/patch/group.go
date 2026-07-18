package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchGroup(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	trueVal := true
	orig, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), &trueVal, &trueVal)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.Group{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if _, err := zc.PutGroup(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("group %q updated in domain %s\n", name, domain)
	return nil
}
