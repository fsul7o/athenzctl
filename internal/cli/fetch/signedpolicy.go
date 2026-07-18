package fetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/printer"
)

// newSignedPolicy implements `athenzctl fetch signedpolicy <domain>`.
// It calls ZTS PostSignedPolicyRequest, which returns policy data in JWS
// form (JWSPolicyData) — the modern format that supersedes the older
// DomainSignedPolicyData.
func newSignedPolicy(opts *cliopts.Options) *cobra.Command {
	var (
		outputDir      string
		policyVersion  string
		p1363Signature bool
	)
	cmd := &cobra.Command{
		Use:   "signedpolicy DOMAIN [DOMAIN...]",
		Short: "Fetch JWS-signed policy data from ZTS",
		Long: `Fetch JWS-signed policy data for one or more domains from ZTS.
The response format is a JWS envelope ({payload, protected, header,
signature}) suitable for ZPE-compatible policy engines.

Without --output-dir the JWS document is written to stdout as JSON (one
per domain, separated by newlines). With --output-dir each domain is
written to "<output-dir>/<domain>.pol" atomically. Absent domains are
reported to stderr and the command continues with the rest.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zc, err := opts.ZTSClient()
			if err != nil {
				return err
			}
			format, err := opts.Format()
			if err != nil {
				return err
			}

			if outputDir != "" {
				if err := os.MkdirAll(outputDir, 0o755); err != nil {
					return fmt.Errorf("mkdir output-dir: %w", err)
				}
			}

			req := &zts.SignedPolicyRequest{
				PolicyVersions:       map[string]string{},
				SignatureP1363Format: p1363Signature,
			}
			if policyVersion != "" {
				req.PolicyVersions[policyVersion] = ""
			}

			var lastErr error
			for _, dom := range args {
				jws, _, err := zc.PostSignedPolicyRequest(zts.DomainName(dom), req, "")
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %v\n", dom, cliopts.WrapErr(err))
					lastErr = err
					continue
				}
				if outputDir != "" {
					if err := writePolicyFile(outputDir, dom, jws); err != nil {
						return err
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", filepath.Join(outputDir, dom+".pol"))
					continue
				}
				if err := writeOne(cmd, format, jws); err != nil {
					return err
				}
			}
			if len(args) == 1 && lastErr != nil {
				return cliopts.WrapErr(lastErr)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "write each domain's JWS policy to <output-dir>/<domain>.pol")
	cmd.Flags().StringVar(&policyVersion, "policy-version", "", "request a specific active policy version tag")
	cmd.Flags().BoolVar(&p1363Signature, "p1363", false, "request P1363-format signature instead of ASN.1 DER")
	return cmd
}

func writeOne(cmd *cobra.Command, format printer.Format, jws *zts.JWSPolicyData) error {
	if jws == nil {
		return errors.New("ZTS returned no policy data")
	}
	switch format {
	case printer.FormatYAML:
		return printer.WriteYAML(cmd.OutOrStdout(), jws)
	default:
		return printer.WriteJSON(cmd.OutOrStdout(), jws)
	}
}

func writePolicyFile(dir, domain string, jws *zts.JWSPolicyData) error {
	data, err := json.Marshal(jws)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain+".pol")
	tmp, err := os.CreateTemp(dir, ".pol-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
