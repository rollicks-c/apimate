package apimate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rollicks-c/apimate/internal/client"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func WithJSONReceiver(receiver interface{}) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Req.Header.Set("Content-Type", "application/json")
		ctx.Receiver = func(payload []byte) error {

			if ctx.ConsumeAllPages {
				data, err := client.ParseArrayList(payload)
				if err != nil {
					return err
				}
				payload = data
			}

			if err := json.Unmarshal(payload, receiver); err != nil {
				return err
			}

			return nil
		}
		return nil
	}
}

func WithRawReceiver(receiver *[]byte) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = func(payload []byte) error {
			*receiver = payload
			return nil
		}
		return nil
	}
}

func WithCustomReceiver(receiver client.Receiver) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = receiver
		return nil
	}
}

func WithDefaultRequest() client.RequestOption {
	return func(ctx *client.RequestContext) error {
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, nil)
		if err != nil {
			return err
		}
		ctx.Req = req
		return nil
	}
}

func WithNullReceiver() client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = func(bytes []byte) error {
			return nil
		}
		return nil
	}
}

func WithPayload(body io.Reader) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, body)
		if err != nil {
			return err
		}
		ctx.Req = req
		return nil
	}
}

func WithFormPayload(body io.Reader) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx.Req = req
		return nil
	}
}

func WithJSONPayload(data any) client.RequestOption {
	return func(ctx *client.RequestContext) error {

		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		body := bytes.NewReader(raw)
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		ctx.Req = req
		return nil
	}
}

func WithValues(values url.Values) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, strings.NewReader(values.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx.Req = req
		return nil
	}
}

func WithQuery(values url.Values) client.RequestOption {
	return func(ctx *client.RequestContext) error {

		// parse url
		parsedURL, err := url.Parse(ctx.Endpoint)
		if err != nil {
			return err
		}

		// add values
		parsedURL.RawQuery = values.Encode()
		ctx.Endpoint = parsedURL.String()
		ctx.Req.URL = parsedURL

		return nil
	}
}

func WithAllPages() client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.ConsumeAllPages = true
		return nil
	}
}

func WithAcceptedErrors(code ...int) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.AcceptedErrorCodes = code
		return nil
	}
}

type Client struct {
	apiUrl   string
	apiToken string
}

func New(apiUrl, apiToken string) *Client {
	return &Client{
		apiUrl:   apiUrl,
		apiToken: apiToken,
	}
}

func (c Client) Request(method, ep string, options ...client.RequestOption) error {

	// create context with default options
	ctx := &client.RequestContext{
		ApiUrl:          c.apiUrl,
		ApiToken:        c.apiToken,
		Method:          method,
		Endpoint:        fmt.Sprintf("%s/%s", strings.TrimSuffix(c.apiUrl, "/"), strings.TrimPrefix(ep, "/")),
		AutoThrottle:    true,
		AutoRetries:     3,
		ConsumeAllPages: false,
	}

	// apply defaults options
	defaults := []client.RequestOption{
		WithDefaultRequest(),
		WithNullReceiver(),
	}

	// apply custom options
	options = append(defaults, options...)
	for _, option := range options {
		if err := option(ctx); err != nil {
			return err
		}
	}

	// execute
	runner := client.NewRunner(*ctx)
	if err := runner.DoRequest(); err != nil {
		return err
	}

	return nil

}
