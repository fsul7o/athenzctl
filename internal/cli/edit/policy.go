package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// policyWhitelist enumerates fields writable via PutPolicy. Assertion
// `id` is preserved (not stripped): removing it makes ZMS treat every
// assertion as new, silently dropping existing ones.
var PolicyWhitelist = Whitelist{
	"name":          nil,
	"description":   nil,
	"tags":          nil,
	"version":       nil,
	"active":        nil,
	"caseSensitive": nil,
	"assertions": Whitelist{
		"id":            nil,
		"effect":        nil,
		"action":        nil,
		"resource":      nil,
		"role":          nil,
		"caseSensitive": nil,
		"conditions":    nil,
	},
}

func editPolicy(zc *zms.ZMSClient, domain, name, auditRef string) error {
	orig, err := zc.GetPolicy(zms.DomainName(domain), zms.EntityName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.Policy{}
	if changed, err := editYAML(orig, edited, "policy-"+name, PolicyWhitelist); err != nil || !changed {
		return err
	}
	if _, err := zc.PutPolicy(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("policy %q updated in domain %s\n", name, domain)
	return nil
}
