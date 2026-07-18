package describe

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeGroupMeta(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("describe group-meta requires NAME")
	}
	g, err := zc.GetGroup(zms.DomainName(domain), zms.EntityName(name), nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta := &zms.GroupMeta{}
	if err := jsonRoundTrip(g, meta); err != nil {
		return err
	}
	return render(w, format, meta)
}
