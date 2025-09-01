package apimate

import (
	"fmt"
	"github.com/rollicks-c/apimate/internal/client"
	"net/http"
	"strings"
)

type Option = client.RequestOption
type JsonBool = client.JsonBool
type JsonInt64 = client.JsonInt64

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
		ApiUrl:             c.apiUrl,
		Method:             method,
		Endpoint:           fmt.Sprintf("%s/%s", strings.TrimSuffix(c.apiUrl, "/"), strings.TrimPrefix(ep, "/")),
		AutoThrottle:       true,
		AutoRetries:        3,
		Paging:             client.PagingConfig{ConsumeAll: false},
		DefaultOptions:     c.defaultOptions,
		ResponseProcessors: []client.ResponseProcessor{},
		SkipTLSVerify:      false,
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
