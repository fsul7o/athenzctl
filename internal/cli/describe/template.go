package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeTemplate(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	tpl, err := zc.GetTemplate(zms.SimpleName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, tpl)
}
