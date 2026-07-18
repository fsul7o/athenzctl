package resource

import (
	"fmt"
	"strings"
)

// PolicyVersionRef is a parsed "POLICY[:VERSION]" reference. Version is
// empty when only the policy name was supplied (list-all-versions form).
type PolicyVersionRef struct {
	Policy  string
	Version string
}

// ParsePolicyVersion accepts "POLICY" or "POLICY:VERSION". Extra colons are
// an error since neither Athenz policy names nor version simple-names allow
// colons.
func ParsePolicyVersion(s string) (PolicyVersionRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return PolicyVersionRef{}, fmt.Errorf("policyversion reference is empty")
	}
	switch parts := strings.Split(s, ":"); len(parts) {
	case 1:
		return PolicyVersionRef{Policy: parts[0]}, nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return PolicyVersionRef{}, fmt.Errorf("invalid policyversion reference %q (want POLICY:VERSION)", s)
		}
		return PolicyVersionRef{Policy: parts[0], Version: parts[1]}, nil
	default:
		return PolicyVersionRef{}, fmt.Errorf("invalid policyversion reference %q (want POLICY:VERSION)", s)
	}
}
