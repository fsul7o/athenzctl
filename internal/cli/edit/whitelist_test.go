package edit

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func TestApplyWhitelist_TopLevel(t *testing.T) {
	src := map[string]any{
		"name":     "admins",
		"modified": "2026-01-01T00:00:00Z",
		"selfServe": true,
	}
	wl := Whitelist{"name": nil, "selfServe": nil}
	got, err := applyWhitelist(src, wl)
	if err != nil {
		t.Fatalf("applyWhitelist: %v", err)
	}
	want := map[string]any{"name": "admins", "selfServe": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestApplyWhitelist_NestedMap(t *testing.T) {
	src := map[string]any{
		"name": "read-access",
		"resourceOwnership": map[string]any{"role": "user.owner"},
		"metadata": map[string]any{"description": "readers", "internal": "x"},
	}
	wl := Whitelist{
		"name":     nil,
		"metadata": Whitelist{"description": nil},
	}
	got, err := applyWhitelist(src, wl)
	if err != nil {
		t.Fatalf("applyWhitelist: %v", err)
	}
	want := map[string]any{
		"name":     "read-access",
		"metadata": map[string]any{"description": "readers"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestApplyWhitelist_SliceOfObjects(t *testing.T) {
	src := map[string]any{
		"assertions": []any{
			map[string]any{"id": float64(1), "action": "read", "requestTime": "x"},
			map[string]any{"id": float64(2), "action": "write", "requestTime": "y"},
		},
	}
	wl := Whitelist{
		"assertions": Whitelist{"id": nil, "action": nil},
	}
	got, err := applyWhitelist(src, wl)
	if err != nil {
		t.Fatalf("applyWhitelist: %v", err)
	}
	want := map[string]any{
		"assertions": []any{
			map[string]any{"id": float64(1), "action": "read"},
			map[string]any{"id": float64(2), "action": "write"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestApplyWhitelist_MissingKeysNotAdded(t *testing.T) {
	src := map[string]any{"name": "svc"}
	wl := Whitelist{"name": nil, "description": nil, "hosts": Whitelist{"a": nil}}
	got, err := applyWhitelist(src, wl)
	if err != nil {
		t.Fatalf("applyWhitelist: %v", err)
	}
	want := map[string]any{"name": "svc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestApplyWhitelist_JSONError(t *testing.T) {
	// math.NaN cannot be marshaled to JSON.
	src := map[string]any{"bad": math.NaN()}
	if _, err := applyWhitelist(src, Whitelist{"bad": nil}); err == nil {
		t.Fatal("expected error from json.Marshal on NaN")
	}
}

// sanity: filterNode preserves scalars unchanged.
func TestApplyWhitelist_ScalarPassThrough(t *testing.T) {
	src := map[string]any{"n": float64(5), "s": "x", "b": true}
	wl := Whitelist{"n": nil, "s": nil, "b": nil}
	got, err := applyWhitelist(src, wl)
	if err != nil {
		t.Fatalf("applyWhitelist: %v", err)
	}
	// Round-trip via JSON to compare types safely.
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(src)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("got %s, want %s", gotJSON, wantJSON)
	}
}
