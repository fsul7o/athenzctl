// Package edit implements the `athenzctl edit` verb: fetch a resource,
// open it in $EDITOR as YAML, and PUT it back on save. Each editable
// kind lives in its own file and declares a Whitelist of writable
// fields — anything not on the list is hidden from the editor so users
// don't waste effort on server-managed values that PUT would ignore.
package edit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `edit` command.
func New(opts *cliopts.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit KIND [NAME]",
		Short: "Edit an Athenz resource in $EDITOR (fetch, edit, PUT)",
		Long: `Fetch an Athenz resource as YAML, open it in $EDITOR (or the value
of $ATHENZCTL_EDITOR), and PUT the modified version back to ZMS on save.
Aborts if the file is unchanged or if you exit the editor without saving
valid YAML.

Supported kinds: role, policy, policyversion, service, group, domain-meta,
role-meta, group-meta, quota.

For domain-meta and quota, NAME may be used as the domain instead of -d,
so both forms are equivalent:
  edit domain-meta DOMAIN
  edit domain-meta -d DOMAIN`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			name := ""
			if len(args) == 2 {
				name = args[1]
			}
			zc, err := opts.ZMSClient()
			if err != nil {
				return err
			}
			auditRef, _ := cmd.Flags().GetString("audit-ref")

			switch kind {
			case resource.KindDomainMeta:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return editDomainMeta(zc, dom, auditRef)
			case resource.KindQuota:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				return editQuota(zc, dom, auditRef)
			default:
				if name == "" {
					return fmt.Errorf("`edit %s` requires NAME", kind)
				}
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				switch kind {
				case resource.KindRole:
					return editRole(zc, dom, name, auditRef)
				case resource.KindPolicy:
					return editPolicy(zc, dom, name, auditRef)
				case resource.KindGroup:
					return editGroup(zc, dom, name, auditRef)
				case resource.KindRoleMeta:
					return editRoleMeta(zc, dom, name, auditRef)
				case resource.KindGroupMeta:
					return editGroupMeta(zc, dom, name, auditRef)
				case resource.KindPolicyVersion:
					return editPolicyVersion(zc, dom, name, auditRef)
				case resource.KindService:
					return editService(zc, dom, name, auditRef)
				default:
					return fmt.Errorf("`edit %s` is not supported", kind)
				}
			}
		},
	}
	cmd.Flags().String("audit-ref", "", "audit reference message")
	return cmd
}

// editYAML marshals orig — projected through wl to hide read-only fields —
// to YAML, opens it in the user's editor, and unmarshals the (possibly
// edited) result into out. Returns (changed, err): changed=false with
// err=nil means the user made no changes and the caller should skip
// the PUT.
func editYAML(orig, out any, label string, wl Whitelist) (bool, error) {
	filtered, err := applyWhitelist(orig, wl)
	if err != nil {
		return false, err
	}
	original, err := yaml.Marshal(filtered)
	if err != nil {
		return false, err
	}
	f, err := os.CreateTemp("", "athenzctl-"+label+"-*.yaml")
	if err != nil {
		return false, err
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := f.Write(original); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}

	if err := runEditor(path); err != nil {
		return false, err
	}

	edited, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if bytes.Equal(bytes.TrimSpace(original), bytes.TrimSpace(edited)) {
		fmt.Fprintln(os.Stderr, "edit cancelled: no changes")
		return false, nil
	}
	if err := yaml.Unmarshal(edited, out); err != nil {
		return false, fmt.Errorf("parse edited YAML: %w", err)
	}
	return true, nil
}

// runEditor honors $ATHENZCTL_EDITOR, then $EDITOR, then $VISUAL, then a
// short fallback list. It refuses to pick a random binary from PATH.
func runEditor(path string) error {
	candidates := []string{
		os.Getenv("ATHENZCTL_EDITOR"),
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
	}
	fallback := []string{"vim", "vi", "nano"}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		return runOne(c, path)
	}
	for _, name := range fallback {
		if _, err := exec.LookPath(name); err == nil {
			return runOne(name, path)
		}
	}
	return errors.New("no editor found: set $EDITOR or $ATHENZCTL_EDITOR")
}

func runOne(cmdline, path string) error {
	// Simple shell-style split: first token is the program, remainder are
	// argv, plus the target path. Does not handle quoted args, which is
	// consistent with kubectl edit's behavior.
	parts := splitEditorCmd(cmdline)
	parts = append(parts, path)
	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

func splitEditorCmd(s string) []string {
	// os/exec expects argv split; a common EDITOR value like "code --wait"
	// needs to become ["code", "--wait"].
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	if len(out) == 0 {
		out = []string{s}
	}
	return out
}
