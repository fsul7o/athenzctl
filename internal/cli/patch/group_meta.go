package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchGroupMeta(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	orig := &zms.GroupMeta{}
	if err := jsonRoundTrip(g, orig); err != nil {
		return err
	}
	out := &zms.GroupMeta{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if err := zc.PutGroupMeta(zms.DomainName(domain), zms.EntityName(name), auditRef, "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("group-meta %q updated in domain %s\n", name, domain)
	return nil
}
