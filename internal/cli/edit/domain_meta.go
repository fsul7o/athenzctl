package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// domainMetaWhitelist enumerates fields writable via PutDomainMeta.
// System attributes managed by PutDomainSystemMeta (awsAccountName,
// azureTenant, azureClient, gcpProjectNumber, ypmId) are hidden — set
// them through the system-meta API instead. Note: account,
// azureSubscription, gcpProject are self-service on modern Athenz, so
// they remain visible.
var DomainMetaWhitelist = Whitelist{
	"description":           nil,
	"org":                   nil,
	"enabled":               nil,
	"auditEnabled":          nil,
	"applicationId":         nil,
	"memberExpiryDays":      nil,
	"tokenExpiryMins":       nil,
	"serviceCertExpiryMins": nil,
	"roleCertExpiryMins":    nil,
	"signAlgorithm":         nil,
	"serviceExpiryDays":     nil,
	"groupExpiryDays":       nil,
	"userAuthorityFilter":   nil,
	"tags":                  nil,
	"businessService":       nil,
	"memberPurgeExpiryDays": nil,
	"productId":             nil,
	"featureFlags":          nil,
	"contacts":              nil,
	"environment":           nil,
	"x509CertSignerKeyId":   nil,
	"sshCertSignerKeyId":    nil,
	"slackChannel":          nil,
	"onCall":                nil,
	"account":               nil,
	"azureSubscription":     nil,
	"gcpProject":            nil,
}

func editDomainMeta(zc *zms.ZMSClient, name, auditRef string) error {
	d, err := zc.GetDomain(zms.DomainName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	orig := &zms.DomainMeta{}
	if err := jsonRoundTrip(d, orig); err != nil {
		return err
	}
	edited := &zms.DomainMeta{}
	if changed, err := editYAML(orig, edited, "domain-meta-"+name, DomainMetaWhitelist); err != nil || !changed {
		return err
	}
	if err := zc.PutDomainMeta(zms.DomainName(name), auditRef, "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("domain-meta %q updated\n", name)
	return nil
}
