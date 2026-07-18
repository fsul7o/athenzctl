package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsEmpty(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Contexts) != 0 {
		t.Fatalf("expected empty contexts, got %d", len(cfg.Contexts))
	}
	if cfg.APIVersion == "" || cfg.Kind == "" {
		t.Fatalf("expected apiVersion/kind defaulted, got %+v", cfg)
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	orig := New()
	orig.Upsert(Context{Name: "prod", ZMSURL: "https://zms.example", ZTSURL: "https://zts.example", Cert: "/x/c.pem", Key: "/x/k.pem"})
	orig.CurrentContext = "prod"

	if err := Save(path, orig); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CurrentContext != "prod" {
		t.Fatalf("current-context lost: %+v", loaded)
	}
	got, err := loaded.Current()
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if got.ZMSURL != "https://zms.example" {
		t.Fatalf("zms url mismatch: %q", got.ZMSURL)
	}
}

func TestUpsertReplacesByName(t *testing.T) {
	c := New()
	c.Upsert(Context{Name: "a", ZMSURL: "one"})
	c.Upsert(Context{Name: "a", ZMSURL: "two"})
	if len(c.Contexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(c.Contexts))
	}
	if c.Contexts[0].ZMSURL != "two" {
		t.Fatalf("upsert did not replace: %q", c.Contexts[0].ZMSURL)
	}
}

func TestRemoveClearsCurrent(t *testing.T) {
	c := New()
	c.Upsert(Context{Name: "a"})
	c.CurrentContext = "a"
	if !c.Remove("a") {
		t.Fatal("Remove returned false for existing context")
	}
	if c.CurrentContext != "" {
		t.Fatalf("current-context not cleared: %q", c.CurrentContext)
	}
	if c.Remove("missing") {
		t.Fatal("Remove returned true for missing context")
	}
}

func TestCurrentUnsetError(t *testing.T) {
	c := New()
	if _, err := c.Current(); err == nil {
		t.Fatal("expected error when current-context unset")
	}
	c.CurrentContext = "nope"
	if _, err := c.Current(); err == nil {
		t.Fatal("expected error when current-context points to missing entry")
	}
}
