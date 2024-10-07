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
		ctx.Receiver = func(payload [][]byte) error {

			// empty
			if len(payload) == 0 {
				return nil
			}

			// paged
			if len(payload) > 1 {
				merged := bytes.Join(payload, []byte("\n"))
				data, err := client.ParseArrayList(merged)
				if err != nil {
					return err
				}
				payload = [][]byte{data}
			}

			// decode
			if err := json.Unmarshal(payload[0], receiver); err != nil {
				return err
			}

			return nil
		}
		return nil
	}
}

func WithRawReceiver(receiver *[]byte) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = func(payload [][]byte) error {
			*receiver = client.MergePages(payload)
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

func WithNullReceiver() client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = func(bytes [][]byte) error {
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

func WithAllPages(pageParam, pagesHeader string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Paging.ConsumeAll = true
		ctx.Paging.PageParam = pageParam
		ctx.Paging.PageCountHeader = pagesHeader
		return nil
	}
}

func WithAcceptedErrors(codes ...int) client.RequestOption {

	checker := func(resp *http.Response) bool {
		for _, c := range codes {
			if c == resp.StatusCode {
				return true
			}
		}
		return false
	}
	return func(ctx *client.RequestContext) error {
		ctx.StatusChecker = checker
		return nil
	}
}

func WithStatusChecker(checker client.StatusChecker) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.StatusChecker = checker
		return nil
	}
}

type Client struct {
	apiUrl         string
	defaultOptions []client.RequestOption
}

func New(apiUrl string, defaults ...client.RequestOption) *Client {

	return &Client{
		apiUrl:         apiUrl,
		defaultOptions: defaults,
	}
}

func (c Client) Request(method, ep string, options ...client.RequestOption) error {

	// create context with default options
	ctx := &client.RequestContext{
		ApiUrl:         c.apiUrl,
		Method:         method,
		Endpoint:       fmt.Sprintf("%s/%s", strings.TrimSuffix(c.apiUrl, "/"), strings.TrimPrefix(ep, "/")),
		AutoThrottle:   true,
		AutoRetries:    3,
		Paging:         client.PagingConfig{ConsumeAll: false},
		DefaultOptions: c.defaultOptions,
	}

	// apply defaults options
	defaults := []client.RequestOption{
		WithDefaultRequest(),
		WithNullReceiver(),
		WithAcceptedErrors(),
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
