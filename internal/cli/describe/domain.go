package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeDomain(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	d, err := zc.GetDomain(zms.DomainName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, d)
}
