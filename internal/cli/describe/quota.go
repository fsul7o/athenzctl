package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeQuota(w io.Writer, zc *zms.ZMSClient, domain string, format printer.Format) error {
	q, err := zc.GetQuota(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, q)
}
