package cliopts

import (
	"path/filepath"
	"testing"

	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/resource"
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

func TestResolveResourceDomain(t *testing.T) {
	tests := []struct {
		name       string
		kind       resource.Kind
		domainFlag string
		nameArg    string
		want       string
		wantErr    bool
	}{
		{name: "global domain", kind: resource.KindDomain, want: ""},
		{name: "global template", kind: resource.KindTemplate, want: ""},
		{name: "domain meta uses name", kind: resource.KindDomainMeta, domainFlag: "flag.example", nameArg: "name.example", want: "name.example"},
		{name: "quota falls back to flag", kind: resource.KindQuota, domainFlag: "flag.example", want: "flag.example"},
		{name: "scoped resource uses flag", kind: resource.KindRole, domainFlag: "flag.example", nameArg: "reader", want: "flag.example"},
		{name: "scoped resource requires flag", kind: resource.KindGroup, nameArg: "admins", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{Domain: tt.domainFlag}
			got, err := opts.ResolveResourceDomain(tt.kind, tt.nameArg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveResourceDomain() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ResolveResourceDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}
