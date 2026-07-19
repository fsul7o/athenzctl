//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fsul7o/athenzctl/internal/cli"
)

type e2eCoverage struct {
	mu       sync.Mutex
	enabled  bool
	expected map[string]struct{}
	observed map[string]struct{}
	missing  string
}

var coverage e2eCoverage

var resourceOperations = map[string][]string{
	"athenzctl get": {
		"domain", "domain-meta", "role", "role-meta", "policy", "policyversion",
		"service", "servicekey", "group", "group-meta", "membership", "template",
		"domain-template", "quota",
	},
	"athenzctl describe": {
		"domain", "domain-meta", "role", "role-meta", "policy", "policyversion",
		"service", "servicekey", "group", "group-meta", "membership", "template",
		"domain-template", "quota",
	},
	"athenzctl create": {
		"domain", "role", "policy", "policyversion", "service", "servicekey",
		"group", "membership", "domain-template",
	},
	"athenzctl delete": {
		"domain", "role", "policy", "policyversion", "service", "servicekey",
		"group", "membership", "domain-template", "quota",
	},
	"athenzctl edit": {
		"domain-meta", "quota", "role", "policy", "policyversion", "service",
		"group", "role-meta", "group-meta",
	},
	"athenzctl patch": {
		"domain-meta", "quota", "role", "policy", "policyversion", "service",
		"group", "role-meta", "group-meta",
	},
	"athenzctl lookup": {
		"domain",
	},
}

func initializeCoverage() {
	coverage.mu.Lock()
	defer coverage.mu.Unlock()

	coverage.enabled = coverageEnabled()
	coverage.expected = make(map[string]struct{})
	coverage.observed = make(map[string]struct{})
	coverage.missing = ""
	if !coverage.enabled {
		return
	}

	collectCoverageExpectations(cli.NewRootCmd())
}

func finalizeCoverage() {
	coverage.mu.Lock()
	defer coverage.mu.Unlock()

	if !coverage.enabled {
		return
	}

	missing := make([]string, 0)
	for key := range coverage.expected {
		if _, ok := coverage.observed[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		coverage.missing = strings.Join(missing, "\n")
	}
}

func coverageFailure() string {
	coverage.mu.Lock()
	defer coverage.mu.Unlock()
	return coverage.missing
}

func coverageEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("ATHENZCTL_E2E_COVERAGE")))
	return value == "1" || value == "true" || value == "yes"
}

func collectCoverageExpectations(cmd *cobra.Command) {
	if cmd.Run != nil || cmd.RunE != nil {
		coverage.expected[commandCoverageKey(cmd)] = struct{}{}
	}
	for _, kind := range resourceOperations[cmd.CommandPath()] {
		coverage.expected[resourceCoverageKey(cmd.CommandPath(), kind)] = struct{}{}
	}

	cmd.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
		coverage.expected[flagCoverageKey(cmd.CommandPath(), "--"+flag.Name)] = struct{}{}
		if flag.Shorthand != "" {
			coverage.expected[flagCoverageKey(cmd.CommandPath(), "-"+flag.Shorthand)] = struct{}{}
		}
	})
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		coverage.expected[flagCoverageKey(cmd.CommandPath(), "--"+flag.Name)] = struct{}{}
		if flag.Shorthand != "" {
			coverage.expected[flagCoverageKey(cmd.CommandPath(), "-"+flag.Shorthand)] = struct{}{}
		}
	})

	for _, child := range cmd.Commands() {
		if child.Hidden || !child.IsAvailableCommand() || child.Name() == "help" {
			continue
		}
		collectCoverageExpectations(child)
	}
}

func recordCoverage(root *cobra.Command, args []string) {
	coverage.mu.Lock()
	defer coverage.mu.Unlock()
	if !coverage.enabled {
		return
	}

	target, _, err := root.Find(args)
	if err != nil || target == nil {
		target = root
	}
	coverage.observed[commandCoverageKey(target)] = struct{}{}
	if kind := commandKind(target, args); kind != "" {
		key := resourceCoverageKey(target.CommandPath(), kind)
		if _, expected := coverage.expected[key]; expected {
			coverage.observed[key] = struct{}{}
		}
	}

	for _, spelling := range flagSpellings(args) {
		owner := flagOwner(target, spelling)
		if owner == nil {
			continue
		}
		key := flagCoverageKey(owner.CommandPath(), spelling)
		if _, expected := coverage.expected[key]; expected {
			coverage.observed[key] = struct{}{}
		}
	}
}

func commandCoverageKey(cmd *cobra.Command) string {
	return "command " + cmd.CommandPath()
}

func resourceCoverageKey(commandPath, kind string) string {
	return fmt.Sprintf("operation %s %s", commandPath, kind)
}

func flagCoverageKey(commandPath, spelling string) string {
	return fmt.Sprintf("flag %s %s", commandPath, spelling)
}

func flagOwner(target *cobra.Command, spelling string) *cobra.Command {
	name := strings.TrimPrefix(strings.SplitN(spelling, "=", 2)[0], "--")
	short := ""
	if strings.HasPrefix(spelling, "-") && !strings.HasPrefix(spelling, "--") {
		short = strings.TrimPrefix(name, "-")
		name = ""
	}

	for current := target; current != nil; current = current.Parent() {
		if flagMatches(current.LocalNonPersistentFlags(), name, short) {
			return current
		}
		if flagMatches(current.PersistentFlags(), name, short) {
			return current
		}
	}
	return nil
}

func flagMatches(flags *pflag.FlagSet, name, shorthand string) bool {
	if name != "" {
		return flags.Lookup(name) != nil
	}
	matched := false
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Shorthand == shorthand {
			matched = true
		}
	})
	return matched
}

func flagSpellings(args []string) []string {
	spellings := make([]string, 0)
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "--") && len(arg) > 2 {
			spellings = append(spellings, strings.SplitN(arg, "=", 2)[0])
			continue
		}
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			spellings = append(spellings, strings.SplitN(arg, "=", 2)[0])
		}
	}
	return spellings
}

func commandKind(cmd *cobra.Command, args []string) string {
	for index, arg := range args {
		if arg != cmd.Name() || index+1 >= len(args) {
			continue
		}
		kind := args[index+1]
		if strings.HasPrefix(kind, "-") {
			return ""
		}
		for _, supported := range resourceOperations[cmd.CommandPath()] {
			if kind == supported {
				return kind
			}
		}
	}
	return ""
}
