package client

import (
	"errors"

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
	transport, err := Transport(ctx)
	if err != nil {
		return nil, err
	}
	tlsCfg := transport.TLSClientConfig
	if ctx.ZMSServerName != "" {
		tlsCfg.ServerName = ctx.ZMSServerName
	}
	c := zms.NewClient(ctx.ZMSURL, transport)
	return &c, nil
}
