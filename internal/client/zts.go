package client

import (
	"errors"
	"net/http"

	"github.com/AthenZ/athenz/clients/go/zts"

	"github.com/fsul7o/athenzctl/internal/config"
)

// NewZTSClient returns a ZTS client authenticated via the context's mTLS
// credential.
func NewZTSClient(ctx *config.Context) (*zts.ZTSClient, error) {
	if ctx == nil {
		return nil, errors.New("no context")
	}
	if ctx.ZTSURL == "" {
		return nil, errors.New("context is missing zts-url")
	}
	tlsCfg, err := TLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.ZTSServerName != "" {
		tlsCfg.ServerName = ctx.ZTSServerName
	}
	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: defaultTimeout,
	}
	c := zts.NewClient(ctx.ZTSURL, transport)
	return &c, nil
}
