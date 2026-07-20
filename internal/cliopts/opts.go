// Package cliopts holds process-global flag state and helpers so that leaf
// command packages can consume shared inputs (context, output format,
// current domain) without importing the top-level cli package.
package cliopts

import (
	"errors"
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/athenz/clients/go/zts"
	rdl "github.com/ardielle/ardielle-go/rdl"

	"github.com/fsul7o/athenzctl/internal/client"
	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/printer"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// Options mirrors every persistent flag on the root command.
type Options struct {
	ConfigPath               string
	ContextName              string
	Domain                   string
	Output                   string
	InsecureSkipTLSVerify    bool
	InsecureSkipTLSVerifySet bool
	ProxyURL                 string
	ProxyURLSet              bool
}

// ResolvePath returns the config file path, honoring --config, then
// $ATHENZCTL_CONFIG, then $HOME/.athenzctl/config.yaml.
func (o *Options) ResolvePath() (string, error) {
	if o != nil && o.ConfigPath != "" {
		return o.ConfigPath, nil
	}
	return config.DefaultPath()
}

// LoadContext resolves the current context (respecting --context override).
func (o *Options) LoadContext() (*config.Context, error) {
	cfg, _, err := o.LoadConfig()
	if err != nil {
		return nil, err
	}
	name := o.ContextName
	if name == "" {
		name = cfg.CurrentContext
	}
	if name == "" {
		return nil, errors.New("no context selected; run `athenzctl config use-context <name>` or pass --context")
	}
	ctx := cfg.Find(name)
	if ctx == nil {
		path, _ := o.ResolvePath()
		return nil, fmt.Errorf("context %q not found in %s", name, path)
	}
	return ctx, nil
}

// LoadConfig loads the configuration file without requiring a selected
// context. This lets CSR-only issue commands use built-in defaults when no
// context has been configured yet.
func (o *Options) LoadConfig() (*config.Config, string, error) {
	path, err := o.ResolvePath()
	if err != nil {
		return nil, "", err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

// LoadSelectedContext returns the selected context when one is configured.
// It returns nil without an error when no context is selected.
func (o *Options) LoadSelectedContext() (*config.Context, error) {
	cfg, path, err := o.LoadConfig()
	if err != nil {
		return nil, err
	}
	name := o.ContextName
	if name == "" {
		name = cfg.CurrentContext
	}
	if name == "" {
		return nil, nil
	}
	ctx := cfg.Find(name)
	if ctx == nil {
		return nil, fmt.Errorf("context %q not found in %s", name, path)
	}
	return ctx, nil
}

// ZMSClient builds an authenticated ZMS client for the resolved context.
func (o *Options) ZMSClient() (*zms.ZMSClient, error) {
	ctx, err := o.connectionContext()
	if err != nil {
		return nil, err
	}
	return client.NewZMSClient(ctx)
}

// ZTSClient builds an authenticated ZTS client for the resolved context.
func (o *Options) ZTSClient() (*zts.ZTSClient, error) {
	ctx, err := o.connectionContext()
	if err != nil {
		return nil, err
	}
	return client.NewZTSClient(ctx)
}

// connectionContext returns a copy of the selected context with any
// command-line connection overrides applied. Keeping the copy local ensures
// one invocation cannot mutate the loaded configuration.
func (o *Options) connectionContext() (*config.Context, error) {
	ctx, err := o.LoadContext()
	if err != nil {
		return nil, err
	}
	copy := *ctx
	if o.InsecureSkipTLSVerifySet {
		copy.InsecureSkipTLSVerify = o.InsecureSkipTLSVerify
	}
	if o.ProxyURLSet {
		copy.ProxyURL = o.ProxyURL
	}
	return &copy, nil
}

// RequireDomain returns Options.Domain or an error if it is empty.
func (o *Options) RequireDomain() (string, error) {
	if o.Domain == "" {
		return "", errors.New("missing required flag: -d/--domain")
	}
	return o.Domain, nil
}

// ResolveDomain resolves the effective domain from nameArg (positional argument)
// and the -d/--domain flag. If nameArg is non-empty it takes precedence;
// otherwise the -d flag is used. Returns an error if both are empty.
func (o *Options) ResolveDomain(nameArg string) (string, error) {
	if nameArg != "" {
		return nameArg, nil
	}
	return o.RequireDomain()
}

// ResolveResourceDomain applies the shared get/describe domain convention for
// a resource kind. Domain and template are global resources; domain-meta and
// quota accept NAME as the domain; all other kinds require --domain.
func (o *Options) ResolveResourceDomain(kind resource.Kind, name string) (string, error) {
	switch kind {
	case resource.KindDomain, resource.KindTemplate:
		return "", nil
	case resource.KindDomainMeta, resource.KindQuota:
		return o.ResolveDomain(name)
	default:
		return o.RequireDomain()
	}
}

// Format resolves the -o flag.
func (o *Options) Format() (printer.Format, error) {
	return printer.Parse(o.Output)
}

// Ptr returns a pointer to v. Useful for athenz SDK methods that accept
// pointer-typed flags (e.g. returnObj *bool) and panic on nil.
func Ptr[T any](v T) *T { return &v }

// WrapErr converts an RDL error to a shorter, human-readable message.
func WrapErr(err error) error {
	if err == nil {
		return nil
	}
	var re rdl.ResourceError
	if errors.As(err, &re) {
		return fmt.Errorf("athenz error %d: %s", re.Code, re.Message)
	}
	return err
}
