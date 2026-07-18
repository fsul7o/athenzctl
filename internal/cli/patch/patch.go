// Package patch implements the `athenzctl patch` verb: update specific fields
// of an Athenz resource directly from the command line using KEY=VALUE arguments.
// Unlike `edit` (which opens a full YAML in $EDITOR), `patch` is scriptable and
// non-interactive. Only fields allowed by each resource's whitelist can be set.
package patch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	editcmd "github.com/fsul7o/athenzctl/internal/cli/edit"
	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// New returns the `patch` command.
func New(opts *cliopts.Options) *cobra.Command {
	kindFlags := cliopts.KindFlagSpec{
		Common: []string{"audit-ref"},
		ByKind: map[resource.Kind][]string{
			resource.KindRole:          {},
			resource.KindPolicy:        {},
			resource.KindPolicyVersion: {},
			resource.KindService:       {},
			resource.KindGroup:         {},
			resource.KindDomainMeta:    {},
			resource.KindRoleMeta:      {},
			resource.KindGroupMeta:     {},
			resource.KindQuota:         {},
		},
	}
	cmd := &cobra.Command{
		Use:   "patch KIND [NAME] KEY=VALUE [KEY=VALUE ...]",
		Short: "Update specific fields of an Athenz resource from the command line",
		Long: `Update one or more fields of an Athenz resource without opening an editor.
Fields are specified as KEY=VALUE positional arguments after the resource name.
Values are parsed as YAML scalars (booleans, integers, and strings are auto-detected).
Only fields permitted by each resource kind's whitelist are accepted.

Supported kinds: role, policy, policyversion, service, group, domain-meta,
role-meta, group-meta, quota.

Examples:
  # Update role description and selfServe flag
  athenzctl patch role myrole -d example.domain description="my role" selfServe=true

  # Update domain-meta (NAME or -d can be used as domain)
  athenzctl patch domain-meta example.domain description="updated"

  # Update quota values
  athenzctl patch quota -d example.domain role=100 policy=200`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kindStr := args[0]
			kind, err := resource.Parse(kindStr)
			if err != nil {
				return err
			}
			if err := cliopts.ValidateKindFlags(cmd, "patch", kind, kindFlags); err != nil {
				return err
			}

			// Separate NAME from KEY=VALUE args.
			// For domain-scoped-only resources (domain-meta, quota) the second
			// arg might be a KEY=VALUE if no name is required.
			var name string
			var patchArgs []string

			switch kind {
			case resource.KindDomainMeta, resource.KindQuota:
				// NAME is optional; args[1] might be KEY=VALUE
				if len(args) >= 2 && !strings.Contains(args[1], "=") {
					name = args[1]
					patchArgs = args[2:]
				} else {
					patchArgs = args[1:]
				}
			default:
				if len(args) < 2 {
					return fmt.Errorf("`patch %s` requires NAME", kind)
				}
				name = args[1]
				patchArgs = args[2:]
			}

			if len(patchArgs) == 0 {
				return fmt.Errorf("at least one KEY=VALUE argument is required")
			}

			patches, err := parsePatches(patchArgs)
			if err != nil {
				return err
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
				if err := validatePatches(patches, editcmd.DomainMetaWhitelist); err != nil {
					return err
				}
				return patchDomainMeta(zc, dom, auditRef, patches)
			case resource.KindQuota:
				dom, err := opts.ResolveDomain(name)
				if err != nil {
					return err
				}
				if err := validatePatches(patches, editcmd.QuotaWhitelist); err != nil {
					return err
				}
				return patchQuota(zc, dom, auditRef, patches)
			default:
				if name == "" {
					return fmt.Errorf("`patch %s` requires NAME", kind)
				}
				dom, err := opts.RequireDomain()
				if err != nil {
					return err
				}
				switch kind {
				case resource.KindRole:
					if err := validatePatches(patches, editcmd.RoleWhitelist); err != nil {
						return err
					}
					return patchRole(zc, dom, name, auditRef, patches)
				case resource.KindPolicy:
					if err := validatePatches(patches, editcmd.PolicyWhitelist); err != nil {
						return err
					}
					return patchPolicy(zc, dom, name, auditRef, patches)
				case resource.KindPolicyVersion:
					if err := validatePatches(patches, editcmd.PolicyWhitelist); err != nil {
						return err
					}
					return patchPolicyVersion(zc, dom, name, auditRef, patches)
				case resource.KindService:
					if err := validatePatches(patches, editcmd.ServiceWhitelist); err != nil {
						return err
					}
					return patchService(zc, dom, name, auditRef, patches)
				case resource.KindGroup:
					if err := validatePatches(patches, editcmd.GroupWhitelist); err != nil {
						return err
					}
					return patchGroup(zc, dom, name, auditRef, patches)
				case resource.KindRoleMeta:
					if err := validatePatches(patches, editcmd.RoleMetaWhitelist); err != nil {
						return err
					}
					return patchRoleMeta(zc, dom, name, auditRef, patches)
				case resource.KindGroupMeta:
					if err := validatePatches(patches, editcmd.GroupMetaWhitelist); err != nil {
						return err
					}
					return patchGroupMeta(zc, dom, name, auditRef, patches)
				default:
					return fmt.Errorf("`patch %s` is not supported", kind)
				}
			}
		},
	}
	cmd.Flags().String("audit-ref", "", "audit reference message")
	cliopts.SetKindAwareHelp(cmd, kindFlags)
	return cmd
}

// parsePatches parses KEY=VALUE strings into a map. Values are YAML-decoded
// so that booleans (true/false), integers, and strings are handled naturally.
func parsePatches(args []string) (map[string]any, error) {
	patches := make(map[string]any, len(args))
	for _, arg := range args {
		idx := strings.IndexByte(arg, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid argument %q: expected KEY=VALUE", arg)
		}
		key := arg[:idx]
		rawVal := arg[idx+1:]
		if key == "" {
			return nil, fmt.Errorf("invalid argument %q: key must not be empty", arg)
		}
		var val any
		if err := yaml.Unmarshal([]byte(rawVal), &val); err != nil {
			// Fallback to plain string if YAML parse fails.
			val = rawVal
		}
		patches[key] = val
	}
	return patches, nil
}

// validatePatches returns an error if any key in patches is not present in wl
// (top-level check only — nested objects are passed through as-is).
func validatePatches(patches map[string]any, wl editcmd.Whitelist) error {
	var unknown []string
	for k := range patches {
		if _, ok := wl[k]; !ok {
			unknown = append(unknown, k)
		}
	}
	if len(unknown) > 0 {
		return fmt.Errorf("unknown or read-only field(s): %s", strings.Join(unknown, ", "))
	}
	return nil
}

// applyMerge marshals orig to a JSON map, overlays patches, then unmarshals
// into out. This preserves existing fields while overwriting only the patched keys.
func applyMerge(orig any, patches map[string]any, out any) error {
	raw, err := json.Marshal(orig)
	if err != nil {
		return fmt.Errorf("marshal original: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("unmarshal original: %w", err)
	}
	for k, v := range patches {
		m[k] = v
	}
	merged, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal merged: %w", err)
	}
	if err := json.Unmarshal(merged, out); err != nil {
		return fmt.Errorf("unmarshal merged: %w", err)
	}
	return nil
}
