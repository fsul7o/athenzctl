package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// WritePretty renders v as an indented, label-formatted tree. Field
// discovery goes through JSON marshaling (which retains camelCase names
// from `json:` tags — YAML tags in the Athenz SDK are unnamed and would
// collapse to lowercase), then parsing that JSON as YAML gives an
// ordered node tree. Any field the SDK emits (respecting `omitempty`)
// appears — new upstream fields surface without touching this file.
func WritePretty(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// JSON is a valid YAML subset, so yaml.Unmarshal parses the same bytes.
	var root yaml.Node
	if err := yaml.Unmarshal(b, &root); err != nil {
		return err
	}
	if len(root.Content) == 0 {
		return nil
	}
	return renderNode(w, root.Content[0], 0)
}

func renderNode(w io.Writer, n *yaml.Node, indent int) error {
	switch n.Kind {
	case yaml.MappingNode:
		return renderMapping(w, n, indent)
	case yaml.SequenceNode:
		return renderSequenceBare(w, n, indent)
	case yaml.ScalarNode:
		fmt.Fprintf(w, "%s%s\n", pad(indent), n.Value)
	}
	return nil
}

func renderMapping(w io.Writer, n *yaml.Node, indent int) error {
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := n.Content[i]
		v := n.Content[i+1]
		label := prettyLabel(k.Value)
		switch v.Kind {
		case yaml.ScalarNode:
			fmt.Fprintf(w, "%s%-24s %s\n", pad(indent), label+":", displayScalar(v))
		case yaml.SequenceNode:
			if len(v.Content) == 0 {
				fmt.Fprintf(w, "%s%-24s %s\n", pad(indent), label+":", noneMarker)
				continue
			}
			fmt.Fprintf(w, "%s%s (%d):\n", pad(indent), label, len(v.Content))
			if err := renderSequenceBare(w, v, indent+2); err != nil {
				return err
			}
		case yaml.MappingNode:
			if len(v.Content) == 0 {
				fmt.Fprintf(w, "%s%-24s %s\n", pad(indent), label+":", noneMarker)
				continue
			}
			fmt.Fprintf(w, "%s%s:\n", pad(indent), label)
			if err := renderMapping(w, v, indent+2); err != nil {
				return err
			}
		}
	}
	return nil
}

const noneMarker = "<none>"

// displayScalar returns a human-friendly rendering of a scalar node,
// substituting <none> for empty strings and explicit nulls so unset
// fields remain visible.
func displayScalar(v *yaml.Node) string {
	if v.Tag == "!!null" || v.Value == "" || v.Value == "null" {
		return noneMarker
	}
	return v.Value
}

func renderSequenceBare(w io.Writer, n *yaml.Node, indent int) error {
	for _, item := range n.Content {
		switch item.Kind {
		case yaml.ScalarNode:
			fmt.Fprintf(w, "%s- %s\n", pad(indent), item.Value)
		case yaml.MappingNode:
			// Print first field on the "- " line, remaining fields indented.
			if len(item.Content) >= 2 {
				k, v := item.Content[0], item.Content[1]
				label := prettyLabel(k.Value)
				if v.Kind == yaml.ScalarNode {
					fmt.Fprintf(w, "%s- %s: %s\n", pad(indent), label, v.Value)
				} else {
					fmt.Fprintf(w, "%s- %s:\n", pad(indent), label)
					if err := renderNode(w, v, indent+4); err != nil {
						return err
					}
				}
				// remaining pairs
				if len(item.Content) > 2 {
					rest := &yaml.Node{Kind: yaml.MappingNode, Content: item.Content[2:]}
					if err := renderMapping(w, rest, indent+2); err != nil {
						return err
					}
				}
			}
		case yaml.SequenceNode:
			fmt.Fprintf(w, "%s-\n", pad(indent))
			if err := renderSequenceBare(w, item, indent+2); err != nil {
				return err
			}
		}
	}
	return nil
}

func pad(n int) string {
	return strings.Repeat(" ", n)
}

// prettyLabel converts a camelCase / kebab-case key to Title Case with
// spaces, e.g. "memberExpiryDays" -> "Member Expiry Days".
func prettyLabel(k string) string {
	if k == "" {
		return ""
	}
	// Split camelCase (insert space before each uppercase after lower).
	var b strings.Builder
	prev := rune(0)
	for _, r := range k {
		if r == '-' || r == '_' {
			b.WriteRune(' ')
			prev = ' '
			continue
		}
		if unicode.IsUpper(r) && unicode.IsLower(prev) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
		prev = r
	}
	// Capitalize each word.
	words := strings.Fields(b.String())
	for i, w := range words {
		if w == "" {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
