package cliopts

import (
	"path/filepath"
	"testing"

	"github.com/fsul7o/athenzctl/internal/config"
)

func TestConnectionCLIOverridesContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.New()
	cfg.CurrentContext = "prod"
	cfg.Upsert(config.Context{
		Name:                  "prod",
		InsecureSkipTLSVerify: true,
		ProxyURL:              "http://context-proxy.example:8080",
	})
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	opts := &Options{
		ConfigPath:               path,
		InsecureSkipTLSVerify:    false,
		InsecureSkipTLSVerifySet: true,
		ProxyURL:                 "socks5://cli-proxy.example:1080",
		ProxyURLSet:              true,
	}
	ctx, err := opts.connectionContext()
	if err != nil {
		t.Fatal(err)
	}
	if ctx.InsecureSkipTLSVerify || ctx.ProxyURL != "socks5://cli-proxy.example:1080" {
		t.Fatalf("CLI overrides were not applied: %+v", ctx)
	}
}

func TestConnectionCLICanClearContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.New()
	cfg.CurrentContext = "prod"
	cfg.Upsert(config.Context{Name: "prod", InsecureSkipTLSVerify: true, ProxyURL: "socks5://context-proxy.example:1080"})
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	opts := &Options{ConfigPath: path, ProxyURLSet: true, InsecureSkipTLSVerifySet: true}
	ctx, err := opts.connectionContext()
	if err != nil {
		t.Fatal(err)
	}
	if ctx.InsecureSkipTLSVerify || ctx.ProxyURL != "" {
		t.Fatalf("context settings were not cleared: %+v", ctx)
	}
}
