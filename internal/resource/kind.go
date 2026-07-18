// Package resource defines the canonical Athenz resource kinds recognized
// by the verb-resource style commands (get / describe / create / delete /
// edit) and their name aliases.
//
// Ephemeral credentials (access token, service cert, role cert) and cached
// artifacts (signed policy) are NOT resources: they have their own verbs
// (`issue`, `fetch`) with their own dedicated subcommand trees and never
// participate in this registry.
package resource

import (
	"fmt"
	"strings"
)

// Kind is a normalized, singular resource identifier.
type Kind string

const (
	KindDomain         Kind = "domain"
	KindDomainMeta     Kind = "domain-meta"
	KindRole           Kind = "role"
	KindRoleMeta       Kind = "role-meta"
	KindPolicy         Kind = "policy"
	KindPolicyVersion  Kind = "policyversion"
	KindService        Kind = "service"
	KindServiceKey     Kind = "servicekey"
	KindGroup          Kind = "group"
	KindGroupMeta      Kind = "group-meta"
	KindMembership     Kind = "membership"
	KindTemplate       Kind = "template"
	KindDomainTemplate Kind = "domain-template"
	KindQuota          Kind = "quota"
)

// aliases maps every user-typeable spelling to its canonical Kind. Add
// short forms (e.g. "po", "svc") here later; for now only singular / plural.
var aliases = map[string]Kind{
	"domain":           KindDomain,
	"domains":          KindDomain,
	"domain-meta":      KindDomainMeta,
	"domainmeta":       KindDomainMeta,
	"role":             KindRole,
	"roles":            KindRole,
	"role-meta":        KindRoleMeta,
	"rolemeta":         KindRoleMeta,
	"policy":           KindPolicy,
	"policies":         KindPolicy,
	"policyversion":    KindPolicyVersion,
	"policyversions":   KindPolicyVersion,
	"pv":               KindPolicyVersion,
	"service":          KindService,
	"services":         KindService,
	"servicekey":       KindServiceKey,
	"servicekeys":      KindServiceKey,
	"sk":               KindServiceKey,
	"template":         KindTemplate,
	"templates":        KindTemplate,
	"domain-template":  KindDomainTemplate,
	"domaintemplate":   KindDomainTemplate,
	"domain-templates": KindDomainTemplate,
	"dt":               KindDomainTemplate,
	"quota":            KindQuota,
	"quotas":           KindQuota,
	"group":            KindGroup,
	"groups":           KindGroup,
	"group-meta":       KindGroupMeta,
	"groupmeta":        KindGroupMeta,
	"membership":       KindMembership,
	"memberships":      KindMembership,
}

// Parse resolves a user-supplied kind name to its canonical Kind.
func Parse(s string) (Kind, error) {
	if k, ok := aliases[strings.ToLower(strings.TrimSpace(s))]; ok {
		return k, nil
	}
	return "", fmt.Errorf("unknown resource kind %q", s)
}

// KnownNames lists all accepted spellings for shell completion / help text.
func KnownNames() []string {
	out := make([]string, 0, len(aliases))
	for k := range aliases {
		out = append(out, k)
	}
	return out
}
