package issue

import (
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

func newAccessToken(opts *cliopts.Options) *cobra.Command {
	var (
		roles       []string
		expiresIn   int
		proxyFor    string
		authzDetail string
	)
	cmd := &cobra.Command{
		Use:   "accesstoken",
		Short: "Request an OAuth2 access token from ZTS",
		Long: `Request an OAuth2 access token from ZTS.

The scope is built from -d/--domain and -r/--role. When -r is omitted the
special scope "<domain>:domain" (all roles) is requested. Output defaults to
just the token string; use -o json/yaml for the full response.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			domain, err := opts.RequireDomain()
			if err != nil {
				return err
			}
			zc, err := opts.ZTSClient()
			if err != nil {
				return err
			}

			scope := buildScope(domain, roles)
			form := url.Values{}
			form.Set("grant_type", "client_credentials")
			form.Set("scope", scope)
			if expiresIn > 0 {
				form.Set("expires_in", strconv.Itoa(expiresIn))
			}
			if proxyFor != "" {
				form.Set("proxy_for_principal", proxyFor)
			}
			if authzDetail != "" {
				form.Set("authorization_details", authzDetail)
			}
			resp, err := zc.PostAccessTokenRequest(zts.AccessTokenRequest(form.Encode()))
			if err != nil {
				return cliopts.WrapErr(err)
			}
			format, err := opts.Format()
			if err != nil {
				return err
			}
			return writeAccessToken(cmd.OutOrStdout(), format, resp)
		},
	}
	cmd.Flags().StringSliceVarP(&roles, "role", "r", nil, "role name(s) within the domain; may be repeated")
	cmd.Flags().IntVar(&expiresIn, "expires-in", 0, "token lifetime in seconds (0 = server default)")
	cmd.Flags().StringVar(&proxyFor, "proxy-for-principal", "", "issue a token proxied on behalf of this principal")
	cmd.Flags().StringVar(&authzDetail, "authorization-details", "", "RFC 9396 authorization_details JSON string")
	return cmd
}

func writeAccessToken(w io.Writer, format printer.Format, resp *zts.AccessTokenResponse) error {
	if handled, err := printer.WriteStructured(w, format, resp); handled || err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, resp.Access_token)
	return err
}

func buildScope(domain string, roles []string) string {
	if len(roles) == 0 {
		return domain + ":domain"
	}
	scopes := make([]string, 0, len(roles))
	for _, r := range roles {
		scopes = append(scopes, domain+":role."+r)
	}
	return strings.Join(scopes, " ")
}
