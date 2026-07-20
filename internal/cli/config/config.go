// Package config implements the `athenzctl config` subtree (kubectl-style
// context management).
package config

import (
	"github.com/spf13/cobra"

	cfg "github.com/fsul7o/athenzctl/internal/config"
)

// Options wires shared root-level flags into the subcommands.
type Options struct {
	ConfigPath *string
}

// New returns the `config` group command.
func New(opts *Options) *cobra.Command {
	if opts == nil {
		opts = &Options{ConfigPath: new(string)}
	}
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Modify athenzctl configuration (contexts, credentials)",
	}
	cmd.AddCommand(newSetContext(opts))
	cmd.AddCommand(newUseContext(opts))
	cmd.AddCommand(newGetContexts(opts))
	cmd.AddCommand(newCurrentContext(opts))
	cmd.AddCommand(newDeleteContext(opts))
	cmd.AddCommand(newView(opts))
	return cmd
}

func resolvePath(opts *Options) (string, error) {
	if opts != nil && opts.ConfigPath != nil && *opts.ConfigPath != "" {
		return *opts.ConfigPath, nil
	}
	return cfg.DefaultPath()
}

func loadConfig(opts *Options) (*cfg.Config, string, error) {
	path, err := resolvePath(opts)
	if err != nil {
		return nil, "", err
	}
	c, err := cfg.Load(path)
	if err != nil {
		return nil, path, err
	}
	return c, path, nil
}
