// Package config manages the athenzctl configuration file
// (~/.athenzctl/config.yaml), modeled after kubeconfig.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDirName  = ".athenzctl"
	DefaultFileName = "config.yaml"
	EnvConfigPath   = "ATHENZCTL_CONFIG"
)

// Config is the root document of ~/.athenzctl/config.yaml.
type Config struct {
	CurrentContext string    `yaml:"current-context,omitempty"`
	Contexts       []Context `yaml:"contexts,omitempty"`
}

// Context bundles a ZMS/ZTS endpoint pair with an mTLS credential.
//
// ZMSServerName / ZTSServerName override the TLS ServerName (both SNI and
// hostname verification) when the URL host does not match a SAN on the
// server certificate — e.g. dialing https://localhost:4443 against a cert
// whose SAN is athenz-zms-server. Leave empty in normal deployments.
type Context struct {
	Name          string         `yaml:"name"`
	ZMSURL        string         `yaml:"zms-url,omitempty"`
	ZTSURL        string         `yaml:"zts-url,omitempty"`
	Cert          string         `yaml:"cert,omitempty"`            // path to PEM
	Key           string         `yaml:"key,omitempty"`             // path to PEM
	CACert        string         `yaml:"ca-cert,omitempty"`         // optional
	ZMSServerName string         `yaml:"zms-server-name,omitempty"` // optional TLS ServerName override
	ZTSServerName string         `yaml:"zts-server-name,omitempty"` // optional TLS ServerName override
	AuthMode      string         `yaml:"auth-mode,omitempty"`       // "" or "mtls" (default), "exec"
	Exec          *ExecConfig    `yaml:"exec,omitempty"`
	IssueDefaults *IssueDefaults `yaml:"issue-defaults,omitempty"`
}

// IssueDefaults contains optional defaults for the credential-issuing
// commands. Service and role certificate defaults are intentionally separate
// because their CSR conventions can differ between Athenz deployments.
type IssueDefaults struct {
	ServiceCert *CertificateDefaults `yaml:"servicecert,omitempty"`
	RoleCert    *CertificateDefaults `yaml:"rolecert,omitempty"`
}

// CertificateDefaults contains optional CSR defaults for one certificate
// issuing command. Spiffe is a pointer so an explicit false is distinguishable
// from an omitted value.
type CertificateDefaults struct {
	DNSDomain                 string `yaml:"dns-domain,omitempty"`
	SubjectCountry            string `yaml:"subj-c,omitempty"`
	SubjectProvince           string `yaml:"subj-p,omitempty"`
	SubjectOrganization       string `yaml:"subj-o,omitempty"`
	SubjectOrganizationalUnit string `yaml:"subj-ou,omitempty"`
	SpiffeTrustDomain         string `yaml:"spiffe-trust-domain,omitempty"`
	Spiffe                    *bool  `yaml:"spiffe,omitempty"`
}

// ExecConfig names an external command that places a fresh Athenz user
// X.509 certificate + key at fixed file paths — the common pattern for
// Athenz-ecosystem credential tools (e.g. ctyano/athenz-user-cert), which
// write cert/key files rather than emitting structured output. athenzctl
// execs Command with Args/Env, and once it exits successfully, reads the
// certificate and key PEM from CertPath/KeyPath.
type ExecConfig struct {
	Command  string            `yaml:"command"`
	Args     []string          `yaml:"args,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	CertPath string            `yaml:"cert-path"` // path Command writes the cert PEM to
	KeyPath  string            `yaml:"key-path"`  // path Command writes the key PEM to
}

// New returns an empty but structurally valid Config.
func New() *Config {
	return &Config{}
}

// DefaultPath returns the configuration file path, honoring $ATHENZCTL_CONFIG
// then falling back to $HOME/.athenzctl/config.yaml.
func DefaultPath() (string, error) {
	if p := os.Getenv(EnvConfigPath); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DefaultDirName, DefaultFileName), nil
}

// Load reads the config file at path. If the file does not exist an empty
// Config is returned along with a nil error, so callers can begin populating
// contexts without a bootstrap step.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return nil, err
	}
	cfg := New()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// Save atomically writes cfg to path, creating parent directories as needed.
func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// Find returns a pointer to the named context, or nil if absent.
func (c *Config) Find(name string) *Context {
	for i := range c.Contexts {
		if c.Contexts[i].Name == name {
			return &c.Contexts[i]
		}
	}
	return nil
}

// Upsert inserts or replaces a context by name.
func (c *Config) Upsert(ctx Context) {
	for i := range c.Contexts {
		if c.Contexts[i].Name == ctx.Name {
			c.Contexts[i] = ctx
			return
		}
	}
	c.Contexts = append(c.Contexts, ctx)
}

// Remove deletes the named context and clears CurrentContext if it matched.
// Returns false if the context did not exist.
func (c *Config) Remove(name string) bool {
	for i := range c.Contexts {
		if c.Contexts[i].Name == name {
			c.Contexts = append(c.Contexts[:i], c.Contexts[i+1:]...)
			if c.CurrentContext == name {
				c.CurrentContext = ""
			}
			return true
		}
	}
	return false
}

// Current returns the currently selected context, or an error if unset /
// missing.
func (c *Config) Current() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, errors.New("no current context set; run `athenzctl config use-context <name>`")
	}
	ctx := c.Find(c.CurrentContext)
	if ctx == nil {
		return nil, fmt.Errorf("current-context %q not found in contexts", c.CurrentContext)
	}
	return ctx, nil
}
