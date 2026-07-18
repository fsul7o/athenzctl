package create

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/spf13/cobra"

	"github.com/fsul7o/athenzctl/internal/cliopts"
	"github.com/fsul7o/athenzctl/internal/resource"
)

// yBase64Chars is the Athenz Y-Base64 alphabet (URL-safe, uses '.' '_' '-').
// See github.com/AthenZ/athenz libs/go/zmscli/ybase64.go.
const yBase64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._"

func yBase64Encode(b []byte) string {
	enc := base64.NewEncoding(yBase64Chars).WithPadding('-')
	return enc.EncodeToString(b)
}

func createServiceKey(cmd *cobra.Command, w io.Writer, zc *zms.ZMSClient, domain, name, auditRef string) error {
	ref, err := resource.ParseServiceKey(name)
	if err != nil {
		return err
	}
	if ref.KeyID == "" {
		return errors.New("create servicekey requires SERVICE:KEYID")
	}
	pemPath, _ := cmd.Flags().GetString("pem")
	inlineKey, _ := cmd.Flags().GetString("key")
	if (pemPath == "") == (inlineKey == "") {
		return errors.New("create servicekey requires exactly one of --pem or --key")
	}

	var encoded string
	if pemPath != "" {
		raw, err := os.ReadFile(pemPath)
		if err != nil {
			return fmt.Errorf("read --pem: %w", err)
		}
		encoded = yBase64Encode(raw)
	} else {
		// A user may paste a raw PEM into --key; detect and re-encode.
		if strings.Contains(inlineKey, "BEGIN") {
			encoded = yBase64Encode([]byte(inlineKey))
		} else {
			encoded = inlineKey
		}
	}

	entry := &zms.PublicKeyEntry{Key: encoded, Id: ref.KeyID}
	if err := zc.PutPublicKeyEntry(zms.DomainName(domain), zms.SimpleName(ref.Service), ref.KeyID, auditRef, "", entry); err != nil {
		return cliopts.WrapErr(err)
	}
	fmt.Fprintf(w, "servicekey %s:%s added in domain %s\n", ref.Service, ref.KeyID, domain)
	return nil
}
