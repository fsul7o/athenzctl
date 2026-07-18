package get

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func getServiceKey(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("get servicekey requires SERVICE[:KEYID]")
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
		if done, err := renderStructured(w, format, pk); done || err != nil {
			return err
		}
		return renderServiceKeyTable(w, ref.Service, []*zms.PublicKeyEntry{pk})
	}
	s, err := zc.GetServiceIdentity(zms.DomainName(domain), zms.SimpleName(ref.Service))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := renderStructured(w, format, s.PublicKeys); done || err != nil {
		return err
	}
	return renderServiceKeyTable(w, ref.Service, s.PublicKeys)
}

func renderServiceKeyTable(w io.Writer, service string, keys []*zms.PublicKeyEntry) error {
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		if k == nil {
			continue
		}
		rows = append(rows, []string{service, k.Id, keyHead(k.Key)})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"SERVICE", "KEY-ID", "KEY-HEAD"},
		Rows:    rows,
	})
}

// keyHead trims a Y-Base64 encoded key to a short prefix for table display.
func keyHead(k string) string {
	const n = 24
	if len(k) <= n {
		return k
	}
	return k[:n] + "..."
}
