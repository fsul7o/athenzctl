package get

import (
	"path"
	"strconv"
	"strings"

	rdl "github.com/ardielle/ardielle-go/rdl"
)

// ts formats an optional rdl.Timestamp, returning "-" when absent.
func ts(t *rdl.Timestamp) string {
	if t == nil {
		return "-"
	}
	return t.String()
}

// countStr returns "-" for 0, otherwise a decimal string. Empty tables read
// more cleanly this way.
func countStr(n int) string {
	if n == 0 {
		return "-"
	}
	return strconv.Itoa(n)
}

// shortName strips the "<domain>:role.<name>" / "<domain>:policy.<name>" /
// "<domain>:group.<name>" prefix, returning the trailing simple name.
func shortName(qualified string) string {
	if i := strings.Index(qualified, ":"); i >= 0 {
		qualified = qualified[i+1:]
	}
	return path.Base(strings.ReplaceAll(qualified, ".", "/"))
}
