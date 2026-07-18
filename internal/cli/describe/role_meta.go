package describe

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeRoleMeta(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("describe role-meta requires NAME")
	}
	r, err := zc.GetRole(zms.DomainName(domain), zms.EntityName(name), nil, nil, nil)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta := &zms.RoleMeta{}
	if err := jsonRoundTrip(r, meta); err != nil {
		return err
	}
	return render(w, format, meta)
}
