// Package ntokenauth implements athenzctl's auth-mode "ntoken": on every
// invocation, a ZMS service token (NToken) is signed with a private key
// already registered in ZMS, and used to authenticate a ZTS
// PostInstanceRefreshRequest call that mints a fresh, short-lived X.509
// service certificate. That certificate (and the same private key) then
// become the mTLS credential used for the invocation's ZMS/ZTS calls.
//
// This mirrors the "Request Service Identity Certificate using Registered
// Public/Private Key Pair" mode of upstream zts-svccert. The resulting
// cert/key pair is cached on disk and reused across invocations until it
// nears expiry, to avoid contacting ZTS on every single athenzctl command.
package ntokenauth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AthenZ/athenz/clients/go/zts"
	"github.com/AthenZ/athenz/libs/go/zmssvctoken"

	"github.com/fsul7o/athenzctl/internal/authcache"
	"github.com/fsul7o/athenzctl/internal/authmode"
	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/csr"
)

// NTokenHeaderDefault is the HTTP header ZTS expects a principal NToken
// under, matching zts-svccert's default -hdr value. It is intentionally a
// var (not a const) so release builds can override it with
// go build -ldflags -X; an empty value falls back to the hardcoded
// "Athenz-Principal-Auth". Per-context override: NTokenAuthConfig.Header
// (yaml `hdr`), which takes priority over this build-time default.
var NTokenHeaderDefault string

// resolveNTokenHeader picks the HTTP header to send the signed NToken
// under, in priority order: context ntoken-auth.hdr (highest) >
// build-time ldflags override (NTokenHeaderDefault) > hardcoded default.
func resolveNTokenHeader(cfg *config.NTokenAuthConfig) string {
	if cfg.Header != "" {
		return cfg.Header
	}
	if NTokenHeaderDefault != "" {
		return NTokenHeaderDefault
	}
	return "Athenz-Principal-Auth"
}

// ntokenExpiration is how long the signed NToken itself is valid for. It is
// only used transiently to authenticate the refresh call, not cached.
const ntokenExpiration = 10 * time.Minute

// Fetch returns a cert/key PEM pair for auth-mode "ntoken". See the package
// doc comment for the overall flow.
func Fetch(ctx *config.Context) (certPEM, keyPEM []byte, err error) {
	cfg := ctx.NTokenAuth
	if cfg == nil {
		return nil, nil, errors.New(`auth-mode is "ntoken" but context has no ntoken-auth config`)
	}
	if cfg.Domain == "" || cfg.Service == "" || cfg.PrivateKeyPath == "" || cfg.KeyID == "" {
		return nil, nil, errors.New("ntoken-auth requires domain, service, private-key, and key-id")
	}

	cacheDir, err := authmode.ResolveCacheDir(ctx, "ntoken", ctx.CacheDir)
	if err != nil {
		return nil, nil, err
	}
	if cachedCert, cachedKey, ok := authcache.Load(cacheDir); ok && authcache.Valid(cachedCert) {
		return cachedCert, cachedKey, nil
	}

	keyBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read ntoken-auth private key: %w", err)
	}

	ntoken, err := SignNToken(cfg.Domain, cfg.Service, cfg.KeyID, keyBytes)
	if err != nil {
		return nil, nil, err
	}

	ztsClient, err := authmode.AnonymousZTSClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	ztsClient.AddCredentials(resolveNTokenHeader(cfg), ntoken)

	signer, err := csr.NewSigner(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("load ntoken-auth private key: %w", err)
	}
	dnsDomain, subjC, subjP, subjO, subjOU, spiffe, spiffeTrustDomain, concatIntermediateCert, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		return nil, nil, err
	}
	if dnsDomain == "" {
		return nil, nil, errors.New("ntoken-auth requires issue-defaults.servicecert.dns-domain to be configured")
	}
	host, commonName, spiffeURI := csr.ServiceIdentityNames(cfg.Domain, cfg.Service, dnsDomain, spiffeTrustDomain, spiffe)
	subj := csr.NewSubject(commonName, subjC, subjP, subjO, subjOU)
	csrPEM, err := csr.GenerateServiceCSR(signer, subj, host, "", "", spiffeURI)
	if err != nil {
		return nil, nil, err
	}

	req := &zts.InstanceRefreshRequest{Csr: csrPEM}
	ident, err := ztsClient.PostInstanceRefreshRequest(zts.CompoundName(cfg.Domain), zts.SimpleName(cfg.Service), req)
	if err != nil {
		return nil, nil, fmt.Errorf("ntoken-auth refresh: %w", err)
	}

	certificate := ident.Certificate
	if concatIntermediateCert {
		certificate += ident.CaCertBundle
	}
	newCertPEM := []byte(certificate)
	if err := authcache.Save(cacheDir, newCertPEM, keyBytes); err != nil {
		return nil, nil, err
	}
	return newCertPEM, keyBytes, nil
}

// SignNToken signs a ZMS service token (NToken) for domain/service with
// keyBytes (a PEM-encoded private key already registered in ZMS as keyID).
// The token is valid for ntokenExpiration and is meant to be used
// immediately to authenticate a single ZTS call, not stored.
func SignNToken(domain, service, keyID string, keyBytes []byte) (string, error) {
	builder, err := zmssvctoken.NewTokenBuilder(domain, service, keyBytes, keyID)
	if err != nil {
		return "", fmt.Errorf("build ntoken: %w", err)
	}
	builder.SetExpiration(ntokenExpiration)
	ntoken, err := builder.Token().Value()
	if err != nil {
		return "", fmt.Errorf("sign ntoken: %w", err)
	}
	return ntoken, nil
}
