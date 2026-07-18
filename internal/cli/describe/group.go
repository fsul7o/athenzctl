package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeGroup(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	trueVal := true
	g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), &trueVal, &trueVal)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, g)
}
