package version

import (
	"bytes"
	"testing"

	buildinfo "github.com/fsul7o/athenzctl/internal/version"
)

func TestNewPrintsBuildInformation(t *testing.T) {
	originalVersion, originalCommit, originalDate := buildinfo.Version, buildinfo.Commit, buildinfo.Date
	buildinfo.Version = "v1.2.3"
	buildinfo.Commit = "abc1234"
	buildinfo.Date = "2026-07-20T00:00:00Z"
	t.Cleanup(func() {
		buildinfo.Version = originalVersion
		buildinfo.Commit = originalCommit
		buildinfo.Date = originalDate
	})

	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	const want = "athenzctl v1.2.3 (commit abc1234, built 2026-07-20T00:00:00Z)\n"
	if got := out.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
