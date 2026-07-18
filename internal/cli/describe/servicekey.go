package describe

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func describeServiceKey(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("describe servicekey requires SERVICE[:KEYID]")
	}
	ref, err := resource.ParseServiceKey(name)
	if err != nil {
		return err
	}
	if ref.KeyID != "" {
		pk, err := zc.GetPublicKeyEntry(zms.DomainName(domain), zms.SimpleName(ref.Service), ref.KeyID)
		if err != nil {
			return cliopts.WrapErr(err)
		}
		return render(w, format, pk)
	}
	s, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(ref.Service))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	return render(w, format, s.PublicKeys)
}
