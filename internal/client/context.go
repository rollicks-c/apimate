package client

import "net/http"

type RequestContext struct {
	ApiUrl   string
	ApiToken string

	Method   string
	Endpoint string

	Req *http.Request

	AutoThrottle       bool
	AutoRetries        int
	ConsumeAllPages    bool
	AcceptedErrorCodes []int
	Receiver           Receiver
}

type RequestOption func(*RequestContext) error

type Receiver func([]byte) error
