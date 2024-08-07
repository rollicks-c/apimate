package client

import (
	"fmt"
	"net/http"
)

type responseProcessor struct {
	ctx RequestContext
}

func (rp responseProcessor) process(resp *http.Response, data []byte) error {

	// check status
	if rp.ctx.StatusChecker(resp) {
		return nil
	}
	if rp.isErrorCode(resp.StatusCode) {
		httpError := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if data != nil {
			httpError = fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(data))
		}
		return httpError
	}

	// process body
	if err := rp.ctx.Receiver(data); err != nil {
		return err
	}

	return nil
}

func (rp responseProcessor) isErrorCode(code int) bool {
	if code < 200 {
		return true
	}
	if code > 299 {
		return true
	}
	return false
}
