package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// groupMeta fields are already *Meta-typed (no server-managed read-only
// fields), so we list every field explicitly and exclude only
// resourceOwnership, which is managed via a dedicated API.
var GroupMetaWhitelist = Whitelist{
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
	"description":             nil,
	"tags":                    nil,
}

func editGroupMeta(zc *zms.ZMSClient, domain, name, auditRef string) error {
	g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	orig := &zms.GroupMeta{}
	if err := jsonRoundTrip(g, orig); err != nil {
		return err
	}
	edited := &zms.GroupMeta{}
	if changed, err := editYAML(orig, edited, "group-meta-"+name, GroupMetaWhitelist); err != nil || !changed {
		return err
	}
	if err := zc.PutGroupMeta(zms.DomainName(domain), zms.EntityName(name), auditRef, "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("group-meta %q updated in domain %s\n", name, domain)
	return nil
}
