package describe

import (
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

// describeDomainTemplate shows the list of templates currently applied to
// the domain. If NAME is provided, the underlying server template's full
// definition is rendered as well (roles/policies/groups/services), so the
// user sees exactly what a given template contributes.
func describeDomainTemplate(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	list, err := zc.GetDomainTemplateList(zms.DomainName(domain))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if name == "" {
		return render(w, format, list)
	}
	tpl, err := zc.GetTemplate(zms.SimpleName(name))
	if err != nil {
		return cliopts.WrapErr(err)
	}
	applied := containsTemplate(list, name)
	return render(w, format, struct {
		Name            string        `json:"name"`
		AppliedToDomain bool          `json:"appliedToDomain"`
		Domain          string        `json:"domain"`
		Template        *zms.Template `json:"template"`
	}{name, applied, domain, tpl})
}

func containsTemplate(list *zms.DomainTemplateList, name string) bool {
	if list == nil {
		return false
	}
	for _, n := range list.TemplateNames {
		if string(n) == name {
			return true
		}
	}
	return false
}
