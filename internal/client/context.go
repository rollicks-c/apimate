package client

import "net/http"

type ResponseProcessor func(*http.Response) error

type PagingConfig struct {
	ConsumeAll      bool
	PageParam       string
	PageCountHeader string
}
type RequestContext struct {
	ApiUrl string

	DefaultOptions []RequestOption

	Method   string
	Endpoint string

	Req *http.Request

	AutoThrottle       bool
	AutoRetries        int
	AcceptedErrorCodes []int
	StatusChecker      StatusChecker
	Receiver           Receiver
	ResponseProcessors []ResponseProcessor
	Paging             PagingConfig
	SkipTLSVerify      bool
}

type RequestOption func(*RequestContext) error

type Receiver func([][]byte) error

type StatusChecker func(resp *http.Response) bool

type CookieReceiver func(cookies []*http.Cookie) error
