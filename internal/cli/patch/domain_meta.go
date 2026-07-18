package patch

import (
	"encoding/json"
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func patchDomainMeta(zc *zms.ZMSClient, domain, auditRef string, patches map[string]any) error {
	d, err := zc.GetDomain(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	// Project Domain → DomainMeta via JSON round-trip (same pattern as editDomainMeta).
	orig := &zms.DomainMeta{}
	if err := jsonRoundTrip(d, orig); err != nil {
		return err
	}
	out := &zms.DomainMeta{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}
	if err := zc.PutDomainMeta(zms.DomainName(domain), auditRef, "", out); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("domain-meta %q updated\n", domain)
	return nil
}

// jsonRoundTrip copies fields from src to dst via JSON marshaling.
func jsonRoundTrip(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
