package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// roleWhitelist enumerates fields writable via PutRole. Excluded:
// modified, auditLog, lastReviewedDate, resourceOwnership (all
// server-managed or set via dedicated APIs). RoleMember membership
// request metadata (requestTime, lastNotifiedTime, requestPrincipal,
// pendingState, systemDisabled, principalType, auditRef) is likewise
// hidden — ZMS ignores those on PutRole.
var RoleWhitelist = Whitelist{
	"name":                    nil,
	"description":             nil,
	"tags":                    nil,
	"selfServe":               nil,
	"memberExpiryDays":        nil,
	"tokenExpiryMins":         nil,
	"certExpiryMins":          nil,
	"signAlgorithm":           nil,
	"serviceExpiryDays":       nil,
	"memberReviewDays":        nil,
	"serviceReviewDays":       nil,
	"reviewEnabled":           nil,
	"notifyRoles":             nil,
	"userAuthorityFilter":     nil,
	"userAuthorityExpiration": nil,
	"groupExpiryDays":         nil,
	"groupReviewDays":         nil,
	"auditEnabled":            nil,
	"deleteProtection":        nil,
	"selfRenew":               nil,
	"selfRenewMins":           nil,
	"maxMembers":              nil,
	"principalDomainFilter":   nil,
	"notifyDetails":           nil,
	"trust":                   nil,
	"members":                 nil, // deprecated flat list
	"roleMembers": Whitelist{
		"memberName":     nil,
		"expiration":     nil,
		"reviewReminder": nil,
		"principalName":  nil,
		"active":         nil,
		"approved":       nil,
	},
}

func editRole(zc *zms.ZMSClient, domain, name, auditRef string) error {
	trueVal := true
	orig, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), &trueVal, nil, &trueVal)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.Role{}
	if changed, err := editYAML(orig, edited, "role-"+name, RoleWhitelist); err != nil || !changed {
		return err
	}
	if _, err := zc.PutRole(zms.DomainName(domain), zms.EntityName(name), auditRef, cliopts.Ptr(false), "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("role %q updated in domain %s\n", name, domain)
	return nil
}
