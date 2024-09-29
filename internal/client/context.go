package client

import "net/http"

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
	Paging             PagingConfig
}

type RequestOption func(*RequestContext) error

type Receiver func([][]byte) error

type StatusChecker func(resp *http.Response) bool
