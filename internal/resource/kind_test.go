package resource

import "testing"

func TestParseCanonical(t *testing.T) {
	cases := map[string]Kind{
		"domain":           KindDomain,
		"Domains":          KindDomain,
		"role":             KindRole,
		"roles":            KindRole,
		"policy":           KindPolicy,
		"policies":         KindPolicy,
		"service":          KindService,
		"services":         KindService,
		"group":            KindGroup,
		"groups":           KindGroup,
		"membership":       KindMembership,
		"memberships":      KindMembership,
		"  ROLES  ":        KindRole,
		"domain-meta":      KindDomainMeta,
		"domainmeta":       KindDomainMeta,
		"role-meta":        KindRoleMeta,
		"rolemeta":         KindRoleMeta,
		"group-meta":       KindGroupMeta,
		"groupmeta":        KindGroupMeta,
		"policyversion":    KindPolicyVersion,
		"policyversions":   KindPolicyVersion,
		"PV":               KindPolicyVersion,
		"servicekey":       KindServiceKey,
		"servicekeys":      KindServiceKey,
		"SK":               KindServiceKey,
		"template":         KindTemplate,
		"templates":        KindTemplate,
		"domain-template":  KindDomainTemplate,
		"domain-templates": KindDomainTemplate,
		"DT":               KindDomainTemplate,
		"quota":            KindQuota,
		"quotas":           KindQuota,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("Parse(%q) = %q, want %q", in, got, want)
		}
	}
	for _, unknown := range []string{"bogus", "accesstoken", "servicecert", "rolecert", "signedpolicy"} {
		if _, err := Parse(unknown); err == nil {
			t.Fatalf("expected error for %q (not a resource kind)", unknown)
		}
	}
}
