package fetch

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/printer"
)

func TestWriteOneUsesJSONForNonStructuredFormats(t *testing.T) {
	tests := []struct {
		name   string
		format printer.Format
	}{
		{name: "default table", format: printer.FormatTable},
		{name: "wide", format: printer.FormatWide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			var out bytes.Buffer
			cmd.SetOut(&out)
			jws := &zts.JWSPolicyData{Payload: "payload", Signature: "signature"}
			if err := writeOne(cmd, tt.format, jws); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), `"payload": "payload"`) {
				t.Fatalf("output = %q, want JSON payload", out.String())
			}
		})
	}
}
