package patch

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchQuota(zc *zms.ZMSClient, domain, auditRef string, patches map[string]any) error {
	orig, err := zc.GetQuota(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.Quota{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	// ZMS validates that the request body's Name matches the URI domain.
	// GetQuota may return an empty Name (server fills it from the URI on read),
	// so re-assert it here.
	out.Name = zms.DomainName(domain)
	if err := zc.PutQuota(zms.DomainName(domain), auditRef, out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("quota updated for domain %s\n", domain)
	return nil
}
