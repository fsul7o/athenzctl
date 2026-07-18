package edit

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

// serviceWhitelist enumerates fields writable via PutServiceIdentity.
// modified and resourceOwnership are excluded. publicKeys are shown for
// reference only — use `servicekey` subcommands to modify them.
var ServiceWhitelist = Whitelist{
	"name":             nil,
	"description":      nil,
	"providerEndpoint": nil,
	"executable":       nil,
	"user":             nil,
	"group":            nil,
	"hosts":            nil,
	"tags":             nil,
	"publicKeys": Whitelist{
		"id":  nil,
		"key": nil,
	},
}

func editService(zc *zms.ZMSClient, domain, name, auditRef string) error {
	orig, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.ServiceIdentity{}
	if changed, err := editYAML(orig, edited, "service-"+name, ServiceWhitelist); err != nil || !changed {
		return err
	}
	if _, err := zc.PutServiceIdentity(zms.DomainName(domain), zms.SimpleName(name), auditRef, cliopts.Ptr(false), "", edited); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Printf("service %q updated in domain %s\n", name, domain)
	return nil
}
