package create

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
)

func createDomainTemplate(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	if name == "" {
		return errors.New("create domain-template requires template NAME")
	}
	paramFlags, _ := cmd.Flags().GetStringSlice("param")
	params := make([]*zms.TemplateParam, 0, len(paramFlags))
	for _, p := range paramFlags {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return fmt.Errorf("invalid --param %q (want KEY=VALUE)", p)
		}
		params = append(params, &zms.TemplateParam{Name: zms.SimpleName(k), Value: v})
	}

	dt := &zms.DomainTemplate{
		TemplateNames: []zms.SimpleName{zms.SimpleName(name)},
		Params:        params,
	}
	if err := zc.PutDomainTemplate(zms.DomainName(domain), auditRef, dt); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "domain-template %q applied to domain %s\n", name, domain)
	return nil
}
