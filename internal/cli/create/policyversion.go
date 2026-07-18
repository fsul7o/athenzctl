package create

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func createPolicyVersion(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	ref, err := resource.ParsePolicyVersion(name)
	if err != nil {
		return err
	}
	if ref.Version == "" {
		return fmt.Errorf("create policyversion requires POLICY:VERSION")
	}
	fromVersion, _ := cmd.Flags().GetString("from-version")

	opts := &zms.PolicyOptions{
		Version:     zms.SimpleName(ref.Version),
		FromVersion: zms.SimpleName(fromVersion),
	}
	if _, err := zc.PutPolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), opts, auditRef, cliopts.Ptr(false), ""); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "policy version %s:%s created in domain %s\n", ref.Policy, ref.Version, domain)
	return nil
}
