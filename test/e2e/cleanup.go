//go:build e2e

package e2e

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/client"
	"github.com/fsul7o/athenzctl/internal/config"
)

// adminPrincipal returns the CN of the mTLS client cert (i.e. the identity
// athenzctl is authenticating as). Domains created during e2e must list this
// principal in --admin-users, otherwise subsequent PutRole/PutPolicy calls
// are rejected with 403.
func adminPrincipal() (string, error) {
	cfg, err := config.Load(os.Getenv("ATHENZCTL_E2E_CONFIG"))
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}
	c, err := cfg.Current()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(c.Cert)
	if err != nil {
		return "", fmt.Errorf("read cert %s: %w", c.Cert, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("no PEM block in %s", c.Cert)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse cert: %w", err)
	}
	return cert.Subject.CommonName, nil
}

// zmsClient builds a ZMS client from the ATHENZCTL_E2E_CONFIG file so cleanup
// helpers can bypass cobra and call ZMS directly.
func zmsClient() (*zms.ZMSClient, error) {
	cfg, err := config.Load(os.Getenv("ATHENZCTL_E2E_CONFIG"))
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	c, err := cfg.Current()
	if err != nil {
		return nil, err
	}
	return client.NewZMSClient(c)
}

// cascadeDeleteDomain removes every child resource (role/service/policy/group)
// before deleting the top-level domain itself. Idempotent — missing resources
// are ignored.
func cascadeDeleteDomain(zc *zms.ZMSClient, domain string) error {
	dn := zms.DomainName(domain)

	if roles, err := zc.GetRoleList(dn, nil, ""); err == nil && roles != nil {
		for _, r := range roles.Names {
			if r == "admin" {
				continue // admin role is auto-managed by the top-level delete
			}
			_ = zc.DeleteRole(dn, zms.EntityName(r), "", "")
		}
	}
	if svcs, err := zc.GetServiceIdentityList(dn, nil, ""); err == nil && svcs != nil {
		for _, s := range svcs.Names {
			_ = zc.DeleteServiceIdentity(dn, zms.SimpleName(s), "", "")
		}
	}
	if pols, err := zc.GetPolicyList(dn, nil, ""); err == nil && pols != nil {
		for _, p := range pols.Names {
			if p == "admin" {
				continue
			}
			_ = zc.DeletePolicy(dn, zms.EntityName(p), "", "")
		}
	}
	if groups, err := zc.GetGroups(dn, nil, "", ""); err == nil && groups != nil {
		for _, g := range groups.List {
			// g.Name is "<domain>:group.<name>"
			short := string(g.Name)
			if i := strings.LastIndex(short, ":group."); i >= 0 {
				short = short[i+len(":group."):]
			}
			_ = zc.DeleteGroup(dn, zms.EntityName(short), "", "")
		}
	}

	if err := zc.DeleteTopLevelDomain(zms.SimpleName(domain), "", ""); err != nil {
		return fmt.Errorf("delete domain %s: %w", domain, err)
	}
	return nil
}

// sweepLeakedDomains removes every top-level domain whose name starts with
// "e2e-" — leftovers from a prior test run that crashed or timed out.
func sweepLeakedDomains() {
	zc, err := zmsClient()
	if err != nil {
		log.Printf("e2e sweep: %v", err)
		return
	}
	list, err := zc.GetDomainList(nil, "", "e2e-", nil, "", nil, "", "", "", "", "", "", "", "", "")
	if err != nil || list == nil {
		return
	}
	swept := 0
	for _, d := range list.Names {
		if err := cascadeDeleteDomain(zc, string(d)); err == nil {
			swept++
		}
	}
	if swept > 0 {
		log.Printf("e2e sweep: removed %d leaked domain(s)", swept)
	}
}
