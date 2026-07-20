// Package copperargosauth implements athenzctl's auth-mode "copperargos":
// a service X.509 certificate is obtained via ZTS's Copper Argos provider
// registration flow (PostInstanceRegisterInformation), authenticated by a
// previously prepared attestation-data file — typically produced out of
// band by `issue instance-register-token --out` (see
// internal/cli/issue/instance_register_token.go), or by a cloud-native
// attestation mechanism. This package does not fetch or refresh the
// attestation-data itself.
//
// Once a certificate has been issued, it is cached and, on later
// invocations, refreshed (PostInstanceRefreshRequest, using the cached
// certificate itself as the mTLS credential) rather than re-registered,
// since most providers only allow registering a given provider/instance
// pair once. Registration is retried from the attestation-data file only
// if no usable cache exists yet, or if a refresh attempt fails.
package copperargosauth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/AthenZ/athenz/clients/go/zts"

	"github.com/fsul7o/athenzctl/internal/authcache"
	"github.com/fsul7o/athenzctl/internal/authmode"
	"github.com/fsul7o/athenzctl/internal/config"
	"github.com/fsul7o/athenzctl/internal/csr"
	"github.com/fsul7o/athenzctl/internal/tlsutil"
)

// Fetch returns a cert/key PEM pair for auth-mode "copperargos". See the
// package doc comment for the overall flow.
func Fetch(ctx *config.Context) (certPEM, keyPEM []byte, err error) {
	cfg := ctx.CopperArgosAuth
	if cfg == nil {
		return nil, nil, errors.New(`auth-mode is "copperargos" but context has no copperargos-auth config`)
	}
	if cfg.Domain == "" || cfg.Service == "" || cfg.Provider == "" || cfg.Instance == "" || cfg.AttestationDataPath == "" {
		return nil, nil, errors.New("copperargos-auth requires domain, service, provider, instance, and attestation-data-path")
	}

	cacheDir, err := authmode.ResolveCacheDir(ctx, "copperargos", ctx.CacheDir)
	if err != nil {
		return nil, nil, err
	}

	if cachedCert, cachedKey, ok := authcache.Load(cacheDir); ok {
		if authcache.Valid(cachedCert) {
			return cachedCert, cachedKey, nil
		}
		if refreshedCert, err := selfRefresh(ctx, cfg, cachedCert, cachedKey); err == nil {
			if err := authcache.Save(cacheDir, refreshedCert, cachedKey); err != nil {
				return nil, nil, err
			}
			return refreshedCert, cachedKey, nil
		}
		// self-refresh failed (e.g. the cached certificate is no longer
		// accepted); fall through to register from scratch below.
	}

	return register(ctx, cfg, cacheDir)
}

// selfRefresh renews the certificate using the cached cert/key pair itself
// as the mTLS credential, mirroring upstream zts-svccert's self-refresh
// pattern (its `-service-cert` mode).
func selfRefresh(ctx *config.Context, cfg *config.CopperArgosAuthConfig, cachedCert, cachedKey []byte) ([]byte, error) {
	if ctx.ZTSURL == "" {
		return nil, errors.New("context is missing zts-url")
	}
	tlsCfg, err := tlsutil.BaseTLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	clientCert, err := tls.X509KeyPair(cachedCert, cachedKey)
	if err != nil {
		return nil, fmt.Errorf("parse cached cert/key: %w", err)
	}
	tlsCfg.Certificates = []tls.Certificate{clientCert}
	if ctx.ZTSServerName != "" {
		tlsCfg.ServerName = ctx.ZTSServerName
	}
	transport, err := tlsutil.Transport(ctx, tlsCfg)
	if err != nil {
		return nil, err
	}
	ztsClient := zts.NewClient(ctx.ZTSURL, transport)

	signer, err := csr.NewSigner(cachedKey)
	if err != nil {
		return nil, fmt.Errorf("load cached private key: %w", err)
	}
	csrPEM, err := buildCSR(ctx, cfg, signer, "")
	if err != nil {
		return nil, err
	}
	_, _, _, _, _, _, _, concatIntermediateCert, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		return nil, err
	}

	req := &zts.InstanceRefreshRequest{Csr: csrPEM}
	ident, err := ztsClient.PostInstanceRefreshRequest(zts.CompoundName(cfg.Domain), zts.SimpleName(cfg.Service), req)
	if err != nil {
		return nil, fmt.Errorf("copperargos-auth self-refresh: %w", err)
	}
	certificate := ident.Certificate
	if concatIntermediateCert {
		certificate += ident.CaCertBundle
	}
	return []byte(certificate), nil
}

// register issues a brand-new certificate via the Copper Argos provider
// registration flow, using the configured attestation-data file. A fresh
// ephemeral CSR key is generated in memory for every registration attempt
// (never persisted before a certificate is successfully issued) — register
// only runs when no valid cache exists, so there is never an existing key
// worth preserving across attempts.
func register(ctx *config.Context, cfg *config.CopperArgosAuthConfig, cacheDir string) (certPEM, keyPEM []byte, err error) {
	attestationData, err := os.ReadFile(cfg.AttestationDataPath)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"read attestation-data-path %q (prepare it beforehand, e.g. with `issue instance-register-token --out %s`): %w",
			cfg.AttestationDataPath, cfg.AttestationDataPath, err)
	}

	keyBytes, err := csr.GenerateRSAPrivateKeyPEM()
	if err != nil {
		return nil, nil, fmt.Errorf("generate copperargos-auth CSR key: %w", err)
	}
	signer, err := csr.NewSigner(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("load copperargos-auth private key: %w", err)
	}
	instanceURI := fmt.Sprintf("athenz://instanceid/%s/%s", cfg.Provider, cfg.Instance)
	csrPEM, err := buildCSR(ctx, cfg, signer, instanceURI)
	if err != nil {
		return nil, nil, err
	}

	ztsClient, err := authmode.AnonymousZTSClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	req := &zts.InstanceRegisterInformation{
		Provider:        zts.ServiceName(cfg.Provider),
		Domain:          zts.DomainName(cfg.Domain),
		Service:         zts.SimpleName(cfg.Service),
		AttestationData: string(attestationData),
		Csr:             csrPEM,
	}
	ident, _, err := ztsClient.PostInstanceRegisterInformation(req)
	if err != nil {
		return nil, nil, fmt.Errorf("copperargos-auth register: %w", err)
	}
	_, _, _, _, _, _, _, concatIntermediateCert, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		return nil, nil, err
	}

	certificate := ident.X509Certificate
	if concatIntermediateCert {
		certificate += ident.X509CertificateSigner
	}
	newCertPEM := []byte(certificate)
	if err := authcache.Save(cacheDir, newCertPEM, keyBytes); err != nil {
		return nil, nil, err
	}
	return newCertPEM, keyBytes, nil
}

// buildCSR generates the CSR shared by register and selfRefresh.
// instanceURI is only meaningful for the initial registration CSR.
func buildCSR(ctx *config.Context, cfg *config.CopperArgosAuthConfig, signer *csr.Signer, instanceURI string) (string, error) {
	dnsDomain, subjC, subjP, subjO, subjOU, spiffe, spiffeTrustDomain, _, err := authmode.ServiceCertDefaults(ctx)
	if err != nil {
		return "", err
	}
	if dnsDomain == "" {
		return "", errors.New("copperargos-auth requires issue-defaults.servicecert.dns-domain to be configured")
	}
	host, commonName, spiffeURI := csr.ServiceIdentityNames(cfg.Domain, cfg.Service, dnsDomain, spiffeTrustDomain, spiffe)
	subj := csr.NewSubject(commonName, subjC, subjP, subjO, subjOU)
	return csr.GenerateServiceCSR(signer, subj, host, instanceURI, "", spiffeURI)
}
