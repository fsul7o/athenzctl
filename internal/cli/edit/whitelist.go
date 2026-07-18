package edit

import (
	"encoding/json"
	"fmt"
)

// Whitelist enumerates the JSON keys retained when marshaling an object
// into the edit YAML. Any field not present here is stripped from the
// editor view — this keeps read-only server fields (modified, auditLog,
// lastReviewedDate, resourceOwnership, membership request metadata, ...)
// out of the user's sight and, since PUT payloads without those keys are
// ignored by ZMS, preserves correctness.
//
// A key mapped to nil is a leaf (keep the scalar/object as-is). A key
// mapped to a non-nil child Whitelist is applied recursively to the
// value at that key — including to every element when the value is a
// slice of objects (e.g. roleMembers[]).
type Whitelist map[string]Whitelist

// applyWhitelist projects v through JSON and returns a map/slice tree
// containing only whitelisted keys. Result is safe to yaml.Marshal.
//
// Keys listed in wl but absent from v are not added. Nested Whitelists
// are applied to child objects and to each element of child arrays.
func applyWhitelist(v any, wl Whitelist) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal for whitelist: %w", err)
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("unmarshal for whitelist: %w", err)
	}
	return filterNode(decoded, wl), nil
}

func filterNode(node any, wl Whitelist) any {
	switch n := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(wl))
		for key, child := range wl {
			val, ok := n[key]
			if !ok {
				continue
			}
			if child == nil {
				out[key] = val
				continue
			}
			out[key] = filterNode(val, child)
		}
		return out
	case []any:
		out := make([]any, 0, len(n))
		for _, item := range n {
			out = append(out, filterNode(item, wl))
		}
		return out
	default:
		return node
	}
}
