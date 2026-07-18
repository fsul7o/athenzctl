package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describePolicy(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	p, err := zc.GetPolicy(zms.DomainName(domain), zms.EntityName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, p)
}
