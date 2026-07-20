package issue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AthenZ/athenz/clients/go/zts"

	"github.com/fsul7o/athenzctl/internal/printer"
)

func TestWriteAccessTokenUsesTokenForNonStructuredFormats(t *testing.T) {
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
			resp := &zts.AccessTokenResponse{Access_token: "token-value", Token_type: "Bearer"}
			if err := writeAccessToken(&out, tt.format, resp); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != "token-value\n" {
				t.Fatalf("output = %q, want token only", got)
			}
			if strings.Contains(out.String(), "access_token") {
				t.Fatalf("output = %q, must not use structured encoding", out.String())
			}
		})
	}
}
