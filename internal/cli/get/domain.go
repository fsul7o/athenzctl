package get

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

const defaultDomainListLimit int32 = 100

func getDomain(cmd interface {
	InOrStdin() io.Reader
	ErrOrStderr() io.Writer
}, w io.Writer, zc *zms.ZMSClient, name string, format printer.Format, all, yes bool) error {
	if name != "" {
		if all || yes {
			return fmt.Errorf("--all/--yes can only be used when listing domains")
		}
		d, err := zc.GetDomain(zms.DomainName(name))
		if err != nil {
			return cliopts.WrapErr(err)
		}
		if done, err := printer.WriteStructured(w, format, d); done || err != nil {
			return err
		}
		return printer.WriteTable(w, printer.Table{
			Headers: []string{"NAME", "MODIFIED", "DESCRIPTION"},
			Rows:    [][]string{{string(d.Name), ts(d.Modified), d.Description}},
		})
	}
	if yes && !all {
		return fmt.Errorf("--yes requires --all")
	}
	if all {
		if err := confirmAllDomains(cmd, yes); err != nil {
			return err
		}
	}
	list, err := getDomainList(zc, all)
	if err != nil {
		return cliopts.WrapErr(err)
	}
	if done, err := printer.WriteStructured(w, format, list); done || err != nil {
		return err
	}
	rows := make([][]string, 0, len(list.Names))
	for _, n := range list.Names {
		rows = append(rows, []string{string(n)})
	}
	return printer.WriteTable(w, printer.Table{Headers: []string{"NAME"}, Rows: rows})
}

func getDomainList(zc *zms.ZMSClient, all bool) (*zms.DomainList, error) {
	limit := defaultDomainListLimit
	list, err := zc.GetDomainList(&limit, "", "", nil, "", nil, "", "", "", "", "", "", "", "", "")
	if err != nil || !all || list == nil || list.Next == "" {
		return list, err
	}

	names := append([]zms.DomainName(nil), list.Names...)
	next := list.Next
	for next != "" {
		page, err := zc.GetDomainList(&limit, next, "", nil, "", nil, "", "", "", "", "", "", "", "", "")
		if err != nil {
			return nil, err
		}
		if page == nil {
			return nil, fmt.Errorf("ZMS returned an empty domain list page")
		}
		names = append(names, page.Names...)
		if page.Next == next {
			return nil, fmt.Errorf("ZMS returned a repeated domain list cursor")
		}
		next = page.Next
	}
	list.Names = names
	list.Next = ""
	return list, nil
}

func confirmAllDomains(cmd interface {
	InOrStdin() io.Reader
	ErrOrStderr() io.Writer
}, yes bool) error {
	if yes {
		return nil
	}

	in := cmd.InOrStdin()
	if file, ok := in.(*os.File); ok {
		info, err := file.Stat()
		if err != nil || info.Mode()&os.ModeCharDevice == 0 {
			return fmt.Errorf("get domains --all requires confirmation; use --yes for non-interactive execution")
		}
	}

	errOut := cmd.ErrOrStderr()
	fmt.Fprintln(errOut, "Warning: this command will retrieve all domains and may return a large result.")
	fmt.Fprint(errOut, "Continue? [y/N]: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return fmt.Errorf("confirmation required; use --yes to skip the prompt")
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("operation cancelled")
	}
}
