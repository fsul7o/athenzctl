package patch

import (
	"errors"
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func patchPolicyVersion(zc *zms.ZMSClient, domain, name, auditRef string, patches map[string]any) error {
	ref, err := resource.ParsePolicyVersion(name)
	if err != nil {
		return err
	}
	if ref.Version == "" {
		return errors.New("patch policyversion requires POLICY:VERSION")
	}
	orig, err := zc.GetPolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), zms.SimpleName(ref.Version))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	out := &zms.Policy{}
	if err := applyMerge(orig, patches, out); err != nil {
		return err
	}

	wasActive := orig.Active != nil && *orig.Active
	isActive := out.Active != nil && *out.Active

	// Preserve server-assigned name and version.
	out.Name = orig.Name
	out.Version = orig.Version

	if _, err := zc.PutPolicy(zms.DomainName(domain), zms.EntityName(ref.Policy), auditRef, cliopts.Ptr(false), "", out); err != nil {
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
