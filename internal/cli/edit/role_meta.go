package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// roleMeta fields are already *Meta-typed (no server-managed read-only
// fields), so we list every field explicitly and exclude only
// resourceOwnership, which is managed via a dedicated API.
var RoleMetaWhitelist = Whitelist{
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
	"description":             nil,
	"tags":                    nil,
}

func editRoleMeta(zc *zms.ZMSClient, domain, name, auditRef string) error {
	r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), nil, nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	orig := &zms.RoleMeta{}
	if err := jsonRoundTrip(r, orig); err != nil {
		return err
	}
	edited := &zms.RoleMeta{}
	if changed, err := editYAML(orig, edited, "role-meta-"+name, RoleMetaWhitelist); err != nil || !changed {
		return err
	}
	if err := zc.PutRoleMeta(zms.DomainName(domain), zms.EntityName(name), auditRef, "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("role-meta %q updated in domain %s\n", name, domain)
	return nil
}
