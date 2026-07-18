package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeService(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	s, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, s)
}
