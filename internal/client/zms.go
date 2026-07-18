package client

import (
	"errors"
	"net/http"

	"github.com/AthenZ/athenz/clients/go/zms"

	"github.com/fsul7o/athenzctl/internal/config"
)

// NewZMSClient returns a ZMS client authenticated via the context's mTLS
// credential.
func NewZMSClient(ctx *config.Context) (*zms.ZMSClient, error) {
	if ctx == nil {
		return nil, errors.New("no context")
	}
	if ctx.ZMSURL == "" {
		return nil, errors.New("context is missing zms-url")
	}
	tlsCfg, err := TLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	if ctx.ZMSServerName != "" {
		tlsCfg.ServerName = ctx.ZMSServerName
	}
	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ResponseHeaderTimeout: defaultTimeout,
	}
	c := zms.NewClient(ctx.ZMSURL, transport)
	return &c, nil
}
