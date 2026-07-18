package edit

import (
	"errors"
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// editPolicyVersion edits a single policy version. Assertion/description
// changes go through PutPolicy; flipping `active: true` (from a previously
// non-active version) calls SetActivePolicyVersion, since ZMS treats the
// active pointer as a separate operation. Both may fire in the same edit.
func editPolicyVersion(zc *zms.ZMSClient, domain, name, auditRef string) error {
	ref, err := resource.ParsePolicyVersion(name)
	if err != nil {
		return err
	}
	if ref.Version == "" {
		return errors.New("edit policyversion requires POLICY:VERSION")
	}
	orig, err := zc.GetPolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), zms.SimpleName(ref.Version))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	edited := &zms.Policy{}
	if changed, err := editYAML(orig, edited, "policyversion-"+ref.Policy+"-"+ref.Version, PolicyWhitelist); err != nil || !changed {
		return err
	}

	wasActive := orig.Active != nil && *orig.Active
	isActive := edited.Active != nil && *edited.Active

	// PutPolicy on the fetched version replaces its assertions/description.
	// ZMS routes by the Version field in the body, so preserve it.
	edited.Name = orig.Name
	edited.Version = orig.Version
	if _, err := zc.PutPolicy(zms.DomainName(domain), zms.EntityName(ref.Policy), auditRef, cliopts.Ptr(false), "", edited); err != nil {
		return cliopts.WrapErr(err)
	}

	if !wasActive && isActive {
		opts := &zms.PolicyOptions{Version: zms.SimpleName(ref.Version)}
		if err := zc.SetActivePolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), opts, auditRef, ""); err != nil {
			return cliopts.WrapErr(err)
		}
		fmt.Printf("policy version %s:%s updated and set active in domain %s\n", ref.Policy, ref.Version, domain)
		return nil
	}
	fmt.Printf("policy version %s:%s updated in domain %s\n", ref.Policy, ref.Version, domain)
	return nil
}
