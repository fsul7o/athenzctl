package describe

import (
	"encoding/json"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func describeDomainMeta(w io.Writer, zc *zms.ZMSClient, name string, format printer.Format) error {
	d, err := zc.GetDomain(zms.DomainName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	meta := &zms.DomainMeta{}
	if err := jsonRoundTrip(d, meta); err != nil {
		return err
	}
	return render(w, format, meta)
}

// jsonRoundTrip projects src onto dst via JSON. Domain / Role / Group share
// JSON tags with their *Meta counterparts, so this exactly extracts the
// meta subset without hand-copying dozens of fields.
func jsonRoundTrip(src, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
