package cliopts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fsul7o/athenzctl/internal/resource"
)

// KindFlagSpec describes flags shared by a command and flags available for
// each supported resource kind.
type KindFlagSpec struct {
	Common []string
	ByKind map[resource.Kind][]string
}

// SetKindAwareHelp limits local flags shown by <verb> <kind> --help. The
// original Cobra help renderer is retained for commands without a known kind.
func SetKindAwareHelp(cmd *cobra.Command, spec KindFlagSpec) {
	helpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		kind, ok := commandKind(c)
		if !ok || !spec.hasKind(kind) {
			helpFunc(c, args)
			return
		}

		restore := hideUnusableFlags(c, spec.allowed(kind))
		defer restore()
		helpFunc(c, args)
	})
}

// ValidateKindFlags rejects explicitly supplied local flags that do not apply
// to kind. Persistent global flags remain available for every kind.
func ValidateKindFlags(cmd *cobra.Command, verb string, kind resource.Kind, spec KindFlagSpec) error {
	if !spec.hasKind(kind) {
		return nil
	}

	allowed := spec.allowed(kind)
	localFlags := cmd.NonInheritedFlags()
	var invalid []string
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if localFlags.Lookup(flag.Name) == nil || allowed[flag.Name] || isCobraFlag(flag.Name) {
			return
		}
		invalid = append(invalid, "--"+flag.Name)
	})
	if len(invalid) == 0 {
		return nil
	}
	sort.Strings(invalid)
	return fmt.Errorf("%s %s does not support flag(s): %s", verb, kind, strings.Join(invalid, ", "))
}

func commandKind(cmd *cobra.Command) (resource.Kind, bool) {
	args := cmd.Flags().Args()
	if len(args) == 0 {
		return "", false
	}
	kind, err := resource.Parse(args[0])
	if err != nil {
		return "", false
	}
	return kind, true
}

func (s KindFlagSpec) hasKind(kind resource.Kind) bool {
	_, ok := s.ByKind[kind]
	return ok
}

func (s KindFlagSpec) allowed(kind resource.Kind) map[string]bool {
	allowed := make(map[string]bool, len(s.Common)+len(s.ByKind[kind])+2)
	for _, name := range s.Common {
		allowed[name] = true
	}
	for _, name := range s.ByKind[kind] {
		allowed[name] = true
	}
	allowed["help"] = true
	allowed["version"] = true
	return allowed
}

func hideUnusableFlags(cmd *cobra.Command, allowed map[string]bool) func() {
	flags := cmd.NonInheritedFlags()
	previous := make(map[string]bool)
	flags.VisitAll(func(flag *pflag.Flag) {
		previous[flag.Name] = flag.Hidden
		flag.Hidden = !allowed[flag.Name]
	})
	return func() {
		flags.VisitAll(func(flag *pflag.Flag) {
			flag.Hidden = previous[flag.Name]
		})
	}
}

func isCobraFlag(name string) bool {
	return name == "help" || name == "version"
}
