//go:build e2e

// Package e2e wires godog step definitions to the in-process athenzctl root
// command. Each scenario constructs a fresh cli.NewRootCmd() so cobra state
// (persistent flag pointers stored in cliopts.Options) is not shared.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/fsul7o/athenzctl/internal/cli"
	"github.com/fsul7o/athenzctl/internal/config"
)

// fakeEditorScript is written to each scenario's tempDir and pointed at via
// $ATHENZCTL_EDITOR. It appends a trailing YAML comment so `athenzctl edit`
// sees a non-empty diff and PUTs the (semantically identical) resource back.
const fakeEditorScript = `#!/bin/sh
printf '\n# e2e-edit-touch\n' >> "$1"
`

type world struct {
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	lastErr error
	domain  string // scenario-scoped unique
	tempDir string
	vars    map[string]string
	ctxName string // per-scenario athenzctl context override
}

func newWorld(t string) *world {
	return &world{
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		vars:   map[string]string{},
	}
}

// run invokes athenzctl in-process with the given args slice, prepending
// --config/--context. Context defaults to "local" but can be overridden by
// $ATHENZCTL_E2E_CONTEXT (whole-run) or by setting w.ctxName (per-scenario,
// via the "using context X" step). stdout/stderr/err are captured on the
// world.
func (w *world) run(args []string) error {
	w.stdout.Reset()
	w.stderr.Reset()
	w.lastErr = nil

	cfg := os.Getenv("ATHENZCTL_E2E_CONFIG")
	ctxName := w.ctxName
	if ctxName == "" {
		ctxName = os.Getenv("ATHENZCTL_E2E_CONTEXT")
	}
	if ctxName == "" {
		ctxName = "local"
	}
	expanded := make([]string, len(args))
	for i, a := range args {
		expanded[i] = w.expand(a)
	}
	full := append([]string{"--config", cfg, "--context", ctxName}, expanded...)

	root := cli.NewRootCmd()
	root.SetOut(w.stdout)
	root.SetErr(w.stderr)
	root.SetArgs(full)
	w.lastErr = root.Execute()
	return nil
}

// runLine tokenizes a single-line command (whitespace split, "..." grouping).
func (w *world) runLine(line string) error {
	args, err := tokenize(w.expand(line))
	if err != nil {
		return err
	}
	return w.run(args)
}

// expand substitutes $DOMAIN / ${DOMAIN} and any world.vars references.
func (w *world) expand(s string) string {
	s = strings.ReplaceAll(s, "$DOMAIN", w.domain)
	s = strings.ReplaceAll(s, "${DOMAIN}", w.domain)
	for k, v := range w.vars {
		s = strings.ReplaceAll(s, "$"+k, v)
	}
	return s
}

// tokenize splits a shell-ish string honoring double quotes only.
func tokenize(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unbalanced quote in %q", s)
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out, nil
}

// ---- Step implementations -------------------------------------------------

func (w *world) freshStack() error {
	// bootstrap script already waited on readiness — smoke a cheap call.
	return w.run([]string{"get", "domain", "sys.auth", "-o", "yaml"})
}

func (w *world) aUniqueDomain(base string) error {
	w.domain = fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
	return nil
}

func (w *world) aDomainExists(name string) error {
	name = w.expand(name)
	admin, err := adminPrincipal()
	if err != nil {
		return err
	}
	if err := w.run([]string{"create", "domain", name, "--admin-users", admin}); err != nil {
		return err
	}
	if w.lastErr != nil {
		return fmt.Errorf("create domain %s: %v: %s", name, w.lastErr, w.stderr.String())
	}
	if w.domain == "" {
		w.domain = name
	}
	return nil
}

func (w *world) aRoleExists(role, domain string) error {
	domain = w.expand(domain)
	admin, err := adminPrincipal()
	if err != nil {
		return err
	}
	if err := w.run([]string{"create", "role", role, "-d", domain, "--members", admin}); err != nil {
		return err
	}
	if w.lastErr != nil {
		return fmt.Errorf("create role %s: %v: %s", role, w.lastErr, w.stderr.String())
	}
	return nil
}

func (w *world) aServiceExists(svc, domain string) error {
	domain = w.expand(domain)
	if err := w.run([]string{"create", "service", svc, "-d", domain}); err != nil {
		return err
	}
	if w.lastErr != nil {
		return fmt.Errorf("create service %s: %v: %s", svc, w.lastErr, w.stderr.String())
	}
	return nil
}

func (w *world) aPolicyExists(name, domain string) error {
	domain = w.expand(domain)
	if err := w.run([]string{"create", "policy", name, "-d", domain}); err != nil {
		return err
	}
	if w.lastErr != nil {
		return fmt.Errorf("create policy %s: %v: %s", name, w.lastErr, w.stderr.String())
	}
	return nil
}

func (w *world) aGroupExists(name, domain string) error {
	domain = w.expand(domain)
	admin, err := adminPrincipal()
	if err != nil {
		return err
	}
	if err := w.run([]string{"create", "group", name, "-d", domain, "--members", admin}); err != nil {
		return err
	}
	if w.lastErr != nil {
		return fmt.Errorf("create group %s: %v: %s", name, w.lastErr, w.stderr.String())
	}
	return nil
}

// prerequisites bootstraps the resources needed for `get/describe <kind>`.
// Values match the Examples tables in features/*.feature.
func (w *world) prerequisitesFor(kind string) error {
	kind = strings.ToLower(kind)
	switch kind {
	case "domain", "domain-meta", "quota", "template", "domain-template":
		return nil // domain itself already exists via Background
	case "role", "role-meta":
		return w.aRoleExists("e2e-role", "$DOMAIN")
	case "service", "servicekey":
		if err := w.aServiceExists("e2e-svc", "$DOMAIN"); err != nil {
			return err
		}
		if kind == "servicekey" {
			// Add a key with id "0" so `get servicekey e2e-svc:0` has something to return.
			return w.run([]string{"create", "servicekey", "e2e-svc:0",
				"-d", w.domain, "--key", generatedTestPublicKey})
		}
		return nil
	case "policy":
		return w.aPolicyExists("e2e-policy", "$DOMAIN")
	case "policyversion":
		if err := w.aPolicyExists("e2e-policy", "$DOMAIN"); err != nil {
			return err
		}
		return w.run([]string{"create", "policyversion", "e2e-policy:v1", "-d", "$DOMAIN",
			"--from-version", "0"})
	case "group", "group-meta":
		return w.aGroupExists("e2e-group", "$DOMAIN")
	case "membership":
		if err := w.aRoleExists("e2e-role", "$DOMAIN"); err != nil {
			return err
		}
		return w.run([]string{"create", "membership", "-d", "$DOMAIN",
			"--role", "e2e-role", "--member", "user.membertest"})
	default:
		return fmt.Errorf("unknown kind for prerequisites: %s", kind)
	}
}

// ---- Assertions -----------------------------------------------------------

func (w *world) shouldSucceed() error {
	if w.lastErr != nil {
		return fmt.Errorf("expected success, got error: %v\nSTDERR:\n%s", w.lastErr, w.stderr.String())
	}
	return nil
}

func (w *world) shouldFailWith(substr string) error {
	if w.lastErr == nil {
		return fmt.Errorf("expected failure containing %q, got success. STDOUT:\n%s", substr, w.stdout.String())
	}
	msg := w.lastErr.Error() + w.stderr.String()
	if !strings.Contains(msg, substr) {
		return fmt.Errorf("expected error to contain %q; got %q", substr, msg)
	}
	return nil
}

func (w *world) stdoutContains(substr string) error {
	if !strings.Contains(w.stdout.String(), substr) {
		return fmt.Errorf("expected stdout to contain %q; got:\n%s", substr, w.stdout.String())
	}
	return nil
}

func (w *world) stdoutIsValid(format string) error {
	body := w.stdout.String()
	switch strings.ToLower(format) {
	case "json":
		var v any
		if err := json.Unmarshal([]byte(body), &v); err != nil {
			return fmt.Errorf("stdout is not valid JSON: %v\n%s", err, body)
		}
	case "yaml":
		var v any
		if err := yaml.Unmarshal([]byte(body), &v); err != nil {
			return fmt.Errorf("stdout is not valid YAML: %v\n%s", err, body)
		}
	default:
		return fmt.Errorf("unknown format %q", format)
	}
	return nil
}

func (w *world) stdoutIsValidPEM() error {
	block, _ := pem.Decode(w.stdout.Bytes())
	if block == nil {
		return fmt.Errorf("stdout is not a PEM block:\n%s", w.stdout.String())
	}
	return nil
}

// ---- Registration ---------------------------------------------------------

// InitializeTestSuite wires suite-scoped hooks: a pre-run sweep that removes
// any leaked "e2e-*" domains from prior interrupted runs so the same athenz
// stack can be reused across many invocations of `make e2e`.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() { sweepLeakedDomains() })
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	var w *world

	ctx.Before(func(c context.Context, sc *godog.Scenario) (context.Context, error) {
		w = newWorld(sc.Name)
		dir, err := os.MkdirTemp("", "athenzctl-e2e-*")
		if err != nil {
			return c, err
		}
		w.tempDir = dir
		// Expose the admin key path (from the loaded config) as $ADMIN_KEY so
		// features can pass an absolute path regardless of test CWD.
		if cfg, err := config.Load(os.Getenv("ATHENZCTL_E2E_CONFIG")); err == nil {
			if cc, err := cfg.Current(); err == nil {
				w.vars["ADMIN_KEY"] = cc.Key
				w.vars["ADMIN_CERT"] = cc.Cert
			}
		}
		// Fake editor for `athenzctl edit` scenarios.
		editorPath := dir + "/fake-editor.sh"
		if err := os.WriteFile(editorPath, []byte(fakeEditorScript), 0o755); err != nil {
			return c, err
		}
		os.Setenv("ATHENZCTL_EDITOR", editorPath)
		return c, nil
	})

	ctx.After(func(c context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if w == nil {
			return c, nil
		}
		os.Unsetenv("ATHENZCTL_EDITOR")
		if w.domain != "" {
			if zc, zerr := zmsClient(); zerr == nil {
				_ = cascadeDeleteDomain(zc, w.domain)
			}
		}
		if w.tempDir != "" {
			_ = os.RemoveAll(w.tempDir)
		}
		return c, nil
	})

	// Givens.
	ctx.Step(`^a fresh athenz stack$`, func() error { return w.freshStack() })
	ctx.Step(`^a unique domain "([^"]+)"$`, func(base string) error { return w.aUniqueDomain(base) })
	ctx.Step(`^a domain "([^"]+)" exists$`, func(name string) error { return w.aDomainExists(name) })
	ctx.Step(`^a role "([^"]+)" exists in domain "([^"]+)"$`, func(r, d string) error { return w.aRoleExists(r, d) })
	ctx.Step(`^a service "([^"]+)" exists in domain "([^"]+)"$`, func(s, d string) error { return w.aServiceExists(s, d) })
	ctx.Step(`^a policy "([^"]+)" exists in domain "([^"]+)"$`, func(p, d string) error { return w.aPolicyExists(p, d) })
	ctx.Step(`^a group "([^"]+)" exists in domain "([^"]+)"$`, func(g, d string) error { return w.aGroupExists(g, d) })
	// ZTS pulls domain updates from ZMS on an interval (see athenz.zts.zms_domain_update_timeout).
	// A short sleep lets fresh domains/services propagate before ZTS-facing commands run.
	ctx.Step(`^ZTS has synced domain "([^"]+)"$`, func(_ string) error { time.Sleep(10 * time.Second); return nil })
	ctx.Step(`^"([^"]+)" prerequisites exist$`, func(kind string) error { return w.prerequisitesFor(kind) })

	// Per-scenario context override.
	ctx.Step(`^I use the "([^"]+)" context$`, func(name string) error { w.ctxName = name; return nil })

	// When.
	ctx.Step(`^I run athenzctl "([^"]*)"$`, func(line string) error { return w.runLine(line) })
	ctx.Step(`^I run athenzctl:$`, func(doc *godog.DocString) error {
		lines := strings.Split(strings.TrimSpace(doc.Content), "\n")
		return w.runLine(strings.Join(lines, " "))
	})

	// Then.
	ctx.Step(`^the command should succeed$`, func() error { return w.shouldSucceed() })
	ctx.Step(`^the command should fail with "([^"]+)"$`, func(s string) error { return w.shouldFailWith(s) })
	ctx.Step(`^stdout should contain "([^"]+)"$`, func(s string) error { return w.stdoutContains(w.expand(s)) })
	ctx.Step(`^stdout should be valid (json|yaml)$`, func(fmt string) error { return w.stdoutIsValid(fmt) })
	ctx.Step(`^stdout should be a valid PEM certificate$`, func() error { return w.stdoutIsValidPEM() })
}

// generatedTestPublicKey is a stable throwaway RSA public key used only to
// register a service key in e2e. Not sensitive.
const generatedTestPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0F3ttrDf9GgJx2wS3vJH
1yEwHF3F0h8XW7EY0KuOw0IIfBHF+DoJqBqK1QVYd0IEcxUlxsGprJqLRomOwYd8
XoQhOO7bcAKtNbjuNn/Ec8AaOJ3Ll8QoRfBRphMpLnHmZDf74HmB1uZuYh0e2mvW
YCf9FRc7uALjO4hDGaSjWZfXk8LPHJTTZ1DfqfEExfJgFiHJUgFOaFbf9pXqDeGh
xf9Qw+9wcm4A/ZubDBSCCTGmz0dtC4Lqb1RM6XjBqoZq6DZQmm8mYZG9J1O88Rlq
5cU2VrpFxD1DnbBEXVSJyJcXe1eGeXcqbtRr8bpqi3sFcSNvsHZ/qXY2/EEZzy14
VwIDAQAB
-----END PUBLIC KEY-----`
