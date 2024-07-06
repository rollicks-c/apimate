package client

import "net/http"

type RequestContext struct {
	ApiUrl string

	DefaultOptions []RequestOption

	Method   string
	Endpoint string

	Req *http.Request

	AutoThrottle       bool
	AutoRetries        int
	ConsumeAllPages    bool
	AcceptedErrorCodes []int
	StatusChecker      StatusChecker
	Receiver           Receiver
}

type RequestOption func(*RequestContext) error

type Receiver func([]byte) error

type StatusChecker func(resp *http.Response) bool
