package resource

import (
	"fmt"
	"strings"
)

// ServiceKeyRef is a parsed "SERVICE[:KEYID]" reference. KeyID is empty when
// only the service name was supplied (list-all-keys form).
type ServiceKeyRef struct {
	Service string
	KeyID   string
}

// ParseServiceKey accepts "SERVICE" or "SERVICE:KEYID". Empty segments or
// extra colons are an error.
func ParseServiceKey(s string) (ServiceKeyRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ServiceKeyRef{}, fmt.Errorf("servicekey reference is empty")
	}
	switch parts := strings.Split(s, ":"); len(parts) {
	case 1:
		return ServiceKeyRef{Service: parts[0]}, nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return ServiceKeyRef{}, fmt.Errorf("invalid servicekey reference %q (want SERVICE:KEYID)", s)
		}
		return ServiceKeyRef{Service: parts[0], KeyID: parts[1]}, nil
	default:
		return ServiceKeyRef{}, fmt.Errorf("invalid servicekey reference %q (want SERVICE:KEYID)", s)
	}
}
