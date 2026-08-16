package describe

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeDomain(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	domains, _, err := zc.GetSignedDomains(
		zms.DomainName(name),
		"false",
		"all",
		boolPtr(true),
		boolPtr(true),
		"",
	)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if domains == nil || len(domains.Domains) != 1 || domains.Domains[0] == nil || domains.Domains[0].Domain == nil {
		return fmt.Errorf("domain with name %s wasn't found", name)
	}
	return render(w, format, domains.Domains[0].Domain)
}

func boolPtr(value bool) *bool { return &value }
