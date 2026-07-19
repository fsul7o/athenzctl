package lookup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
)

// domainSelector contains only the filters accepted by ZMS's domain-list
// lookup API. A selector belongs to exactly one lookup family: role, tag, or
// one of the scalar domain attributes.
type domainSelector struct {
	account         string
	productNumber   *int32
	roleMember      zms.ResourceName
	roleName        zms.ResourceName
	azure           string
	gcp             string
	tagKey          zms.TagKey
	tagValue        zms.TagCompoundValue
	businessService string
	productID       string
}

func parseDomainSelector(raw string) (domainSelector, error) {
	if strings.TrimSpace(raw) == "" {
		return domainSelector{}, fmt.Errorf("missing required flag: --field-selector")
	}

	values := make(map[string]string)
	for _, expression := range strings.Split(raw, ",") {
		expression = strings.TrimSpace(expression)
		if expression == "" {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: empty expression", raw)
		}
		if strings.Contains(expression, "!=") {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: != is not supported", expression)
		}

		parts := strings.SplitN(expression, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: expected key=value", expression)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if strings.HasPrefix(value, "=") {
			value = strings.TrimSpace(value[1:])
		}
		if value == "" {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: value is empty", expression)
		}
		if _, exists := values[key]; exists {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: duplicate key %q", raw, key)
		}
		if !allowedDomainSelectorKey(key) {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: unsupported key %q", raw, key)
		}
		values[key] = value
	}

	if _, ok := values["member"]; ok {
		if _, ok := values["role"]; !ok {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: member requires role", raw)
		}
		if len(values) != 2 {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: role lookup cannot be combined with another lookup", raw)
		}
		return domainSelector{
			roleMember: zms.ResourceName(values["member"]),
			roleName:   zms.ResourceName(values["role"]),
		}, nil
	}
	if _, ok := values["role"]; ok {
		return domainSelector{}, fmt.Errorf("invalid field selector %q: role requires member", raw)
	}

	if _, ok := values["tagValue"]; ok {
		if _, ok := values["tagKey"]; !ok {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: tagValue requires tagKey", raw)
		}
	}
	if _, ok := values["tagKey"]; ok {
		if len(values) == 2 {
			if _, ok := values["tagValue"]; !ok {
				return domainSelector{}, fmt.Errorf("invalid field selector %q: tag lookup cannot be combined with another lookup", raw)
			}
		}
		if len(values) > 2 {
			return domainSelector{}, fmt.Errorf("invalid field selector %q: tag lookup cannot be combined with another lookup", raw)
		}
		return domainSelector{
			tagKey:   zms.TagKey(values["tagKey"]),
			tagValue: zms.TagCompoundValue(values["tagValue"]),
		}, nil
	}

	if len(values) != 1 {
		return domainSelector{}, fmt.Errorf("invalid field selector %q: specify one supported lookup", raw)
	}
	for key, value := range values {
		selector := domainSelector{}
		switch key {
		case "account":
			selector.account = value
		case "azure":
			selector.azure = value
		case "gcp":
			selector.gcp = value
		case "businessService":
			selector.businessService = value
		case "productId":
			if number, err := strconv.ParseInt(value, 10, 32); err == nil {
				productNumber := int32(number)
				selector.productNumber = &productNumber
			} else {
				selector.productID = value
			}
		default:
			return domainSelector{}, fmt.Errorf("invalid field selector %q: unsupported lookup key %q", raw, key)
		}
		return selector, nil
	}

	return domainSelector{}, fmt.Errorf("invalid field selector %q", raw)
}

func allowedDomainSelectorKey(key string) bool {
	switch key {
	case "member", "role", "tagKey", "tagValue", "account", "azure", "gcp", "productId", "businessService":
		return true
	default:
		return false
	}
}
