package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// groupWhitelist enumerates fields writable via PutGroup. Same exclusion
// principle as roleWhitelist.
var GroupWhitelist = Whitelist{
	"name":                    nil,
	"description":             nil,
	"tags":                    nil,
	"selfServe":               nil,
	"reviewEnabled":           nil,
	"notifyRoles":             nil,
	"userAuthorityFilter":     nil,
	"userAuthorityExpiration": nil,
	"auditEnabled":            nil,
	"deleteProtection":        nil,
	"selfRenew":               nil,
	"selfRenewMins":           nil,
	"maxMembers":              nil,
	"principalDomainFilter":   nil,
	"notifyDetails":           nil,
	"memberExpiryDays":        nil,
	"serviceExpiryDays":       nil,
	"groupMembers": Whitelist{
		"memberName": nil,
		"expiration": nil,
		"active":     nil,
		"approved":   nil,
	},
}

func editGroup(zc *zms.ZMSClient, domain, name, auditRef string) error {
	trueVal := true
	orig, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), &trueVal, &trueVal)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.Group{}
	if changed, err := editYAML(orig, edited, "group-"+name, GroupWhitelist); err != nil || !changed {
		return err
	}
	if _, err := zc.PutGroup(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("group %q updated in domain %s\n", name, domain)
	return nil
}
