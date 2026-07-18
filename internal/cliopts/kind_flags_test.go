package cliopts

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/resource"
)

func TestSetKindAwareHelpFiltersFlagsAndRestoresState(t *testing.T) {
	cmd := &cobra.Command{
		Use: "create KIND",
		Run: func(*cobra.Command, []string) {},
	}
	cmd.Flags().String("domain-only", "", "domain flag")
	cmd.Flags().String("membership-only", "", "membership flag")
	cmd.Flags().String("audit-ref", "", "common flag")
	cmd.SetArgs([]string{"domains", "--help"})

	SetKindAwareHelp(cmd, KindFlagSpec{
		Common: []string{"audit-ref"},
		ByKind: map[resource.Kind][]string{
			resource.KindDomain: {"domain-only"},
		},
	})

	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := output.String()
	for _, want := range []string{"--domain-only", "--audit-ref", "--help"} {
		if !strings.Contains(text, want) {
			t.Errorf("help does not contain %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "--membership-only") {
		t.Errorf("help contains filtered flag:\n%s", text)
	}
	if cmd.Flags().Lookup("membership-only").Hidden {
		t.Error("filtered flag remained hidden after help rendering")
	}
}

func TestValidateKindFlags(t *testing.T) {
	newCommand := func() *cobra.Command {
		cmd := &cobra.Command{Use: "create KIND", RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := resource.Parse(args[0])
			if err != nil {
				return err
			}
			return ValidateKindFlags(cmd, "create", kind, KindFlagSpec{
				Common: []string{"audit-ref"},
				ByKind: map[resource.Kind][]string{
					resource.KindDomain: {"domain-only"},
				},
			})
		}}
		cmd.Flags().String("domain-only", "", "domain flag")
		cmd.Flags().String("membership-only", "", "membership flag")
		cmd.Flags().String("audit-ref", "", "common flag")
		return cmd
	}

	t.Run("rejects unsupported flag", func(t *testing.T) {
		cmd := newCommand()
		cmd.SetArgs([]string{"domain", "--membership-only", "value"})
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "--membership-only") {
			t.Fatalf("Execute() error = %v, want unsupported flag error", err)
		}
	})

	t.Run("allows common flag", func(t *testing.T) {
		cmd := newCommand()
		cmd.SetArgs([]string{"domain", "--audit-ref", "ref"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}
