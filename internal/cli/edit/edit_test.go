package edit

import (
	"reflect"
	"testing"
)

func TestSplitEditorCmd(t *testing.T) {
	cases := map[string][]string{
		"vim":                 {"vim"},
		"code --wait":         {"code", "--wait"},
		"  emacs -nw   ":      {"emacs", "-nw"},
		"nvim +startinsert":   {"nvim", "+startinsert"},
	}
	for in, want := range cases {
		got := splitEditorCmd(in)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("splitEditorCmd(%q) = %#v, want %#v", in, got, want)
		}
	}
}
