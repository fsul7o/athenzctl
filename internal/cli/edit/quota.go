package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// quotaWhitelist lists the fields that can be set via PutQuota.
// name and modified are server-managed and excluded.
var QuotaWhitelist = Whitelist{
	"subdomain":   nil,
	"role":        nil,
	"roleMember":  nil,
	"policy":      nil,
	"assertion":   nil,
	"entity":      nil,
	"service":     nil,
	"serviceHost": nil,
	"publicKey":   nil,
	"group":       nil,
	"groupMember": nil,
}

func editQuota(zc *zms.ZMSClient, domain, auditRef string) error {
	orig, err := zc.GetQuota(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.Quota{}
	if changed, err := editYAML(orig, edited, "quota-"+domain, QuotaWhitelist); err != nil || !changed {
		return err
	}
	if err := zc.PutQuota(zms.DomainName(domain), auditRef, edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("quota updated for domain %s\n", domain)
	return nil
}
