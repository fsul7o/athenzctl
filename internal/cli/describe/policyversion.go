package describe

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func describePolicyVersion(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	ref, err := resource.ParsePolicyVersion(name)
	if err != nil {
		return err
	}
	if ref.Version == "" {
		return fmt.Errorf("describe policyversion requires POLICY:VERSION")
	}
	p, err := zc.GetPolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), zms.SimpleName(ref.Version))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, p)
}
