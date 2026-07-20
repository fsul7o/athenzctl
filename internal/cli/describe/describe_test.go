package describe

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fsul7o/athenzctl/internal/printer"
)

func TestRenderUsesPrettyForNonStructuredFormats(t *testing.T) {
	tests := []struct {
		name   string
		format printer.Format
	}{
		{name: "default table", format: printer.FormatTable},
		{name: "wide", format: printer.FormatWide},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := render(&out, tt.format, map[string]string{"displayName": "example"}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "Display Name:") || !strings.Contains(out.String(), "example") {
				t.Fatalf("output = %q, want pretty-rendered value", out.String())
			}
		})
	}
}
