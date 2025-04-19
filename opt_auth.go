package apimate

import (
	"github.com/rollicks-c/apimate/internal/client"
)

func WithBearerAuth(apiToken string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Req.Header.Set("Authorization", "Bearer "+apiToken)
		return nil
	}
}

func WithHeaderAuth(headerKey, token string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Req.Header.Set(headerKey, token)
		return nil
	}
}

func WithBasicAuth(user, pass string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Req.SetBasicAuth(user, pass)
		return nil
	}
}

func WithAuthentikAuth(url, clientID, username, password string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		token, err := client.NewAuthentikAuth(url, clientID, username, password).Authenticate()
		if err != nil {
			return err
		}
		ctx.Req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func WithTlsSkipVerify(skip bool) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.SkipTLSVerify = skip
		return nil
	}
}
