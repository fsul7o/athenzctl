package get

import (
	"fmt"
	"io"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

func getPolicyVersion(w io.Writer, zc *zms.ZMSClient, domain, name string, format printer.Format) error {
	if name == "" {
		return fmt.Errorf("get policyversion requires POLICY[:VERSION]")
	}
	ref, err := resource.ParsePolicyVersion(name)
	if err != nil {
		return err
	}
	if ref.Version != "" {
		p, err := zc.GetPolicyVersion(zms.DomainName(domain), zms.EntityName(ref.Policy), zms.SimpleName(ref.Version))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := renderStructured(w, format, p); done || err != nil {
			return err
		}
		return renderPolicyVersionTable(w, []*zms.Policy{p})
	}
	ps, err := zc.GetPolicies(zms.DomainName(domain), boolPtr(true), boolPtr(true), "", "")
	if err != nil {
		return cliopts.WrapErr(err)
	}
	filtered := filterPolicyVersions(ps.List, ref.Policy)
	if done, err := renderStructured(w, format, filtered); done || err != nil {
		return err
	}
	return renderPolicyVersionTable(w, filtered)
}

// filterPolicyVersions keeps only entries whose short name matches policy.
// GetPolicies returns every policy in the domain; policy names are stored
// qualified as "<domain>:policy.<name>".
func filterPolicyVersions(list []*zms.Policy, policy string) []*zms.Policy {
	out := make([]*zms.Policy, 0, len(list))
	for _, p := range list {
		if p == nil {
			continue
		}
		if shortName(string(p.Name)) == policy {
			out = append(out, p)
		}
	}
	return out
}

func renderPolicyVersionTable(w io.Writer, policies []*zms.Policy) error {
	rows := make([][]string, 0, len(policies))
	for _, p := range policies {
		if p == nil {
			continue
		}
		active := "false"
		if p.Active != nil && *p.Active {
			active = "true"
		}
		ver := string(p.Version)
		if ver == "" {
			ver = "0"
		}
		rows = append(rows, []string{
			shortName(string(p.Name)),
			ver,
			active,
			countStr(len(p.Assertions)),
			ts(p.Modified),
		})
	}
	return printer.WriteTable(w, printer.Table{
		Headers: []string{"POLICY", "VERSION", "ACTIVE", "ASSERTIONS", "MODIFIED"},
		Rows:    rows,
	})
}
