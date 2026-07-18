package printer

import (
	"bytes"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := map[string]Format{
		"":      FormatTable,
		"table": FormatTable,
		"WIDE":  FormatWide,
		"json":  FormatJSON,
		"yaml":  FormatYAML,
		"yml":   FormatYAML,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("Parse(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := Parse("xml"); err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	err := WriteTable(&buf, Table{
		Headers: []string{"NAME", "COUNT"},
		Rows: [][]string{
			{"alpha", "1"},
			{"beta", "22"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestWriteJSONYAML(t *testing.T) {
	obj := map[string]any{"a": 1, "b": "two"}
	var jb, yb bytes.Buffer
	if err := WriteJSON(&jb, obj); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jb.String(), `"a"`) {
		t.Fatalf("json output missing key: %s", jb.String())
	}
	if err := WriteYAML(&yb, obj); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(yb.String(), "a: 1") {
		t.Fatalf("yaml output missing entry: %s", yb.String())
	}
}
