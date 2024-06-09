package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type RequestRunner struct {
	ctx RequestContext
}

type requester func(req *http.Request) (*http.Response, error)

func NewRunner(ctx RequestContext) *RequestRunner {
	return &RequestRunner{
		ctx: ctx,
	}
}

func (r RequestRunner) DoRequest() error {

	// auth
	r.ctx.Req.Header.Set("Authorization", "Bearer "+r.ctx.ApiToken)

	// prepare
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	exe := func(req *http.Request) (*http.Response, error) {
		return client.Do(r.ctx.Req)
	}

	// runner middleware
	exe = r.autoThrottle(exe)
	exe = r.autoRetry(exe)

	// execute
	resp, data, err := r.pagedConsume(exe)
	if err != nil {
		return err
	}

	// process response
	rp := responseProcessor{
		ctx: r.ctx,
	}
	if err := rp.process(resp, data); err != nil {
		return err
	}

	return nil
}

func (r RequestRunner) autoThrottle(rq requester) requester {

	if !r.ctx.AutoThrottle {
		return rq
	}

	at := func(req *http.Request) (*http.Response, error) {
		attempts := 3
		for {

			// run request
			resp, err := rq(req)
			if err != nil {
				return nil, err
			}

			// check rate limit
			if resp.StatusCode == http.StatusTooManyRequests {

				// limit attempts
				attempts--
				if attempts <= 0 {
					return nil, fmt.Errorf("too many attempts")
				}

				// wait and retry
				after := resp.Header.Get("Retry-After")
				seconds, err := strconv.Atoi(after)
				if err != nil {
					seconds = 1
				} else if seconds <= 0 {
					seconds = 1
				}
				time.Sleep(time.Duration(seconds) * time.Second)
				continue

			}

			// done
			return resp, nil
		}
	}

	return at

}

func (r RequestRunner) autoRetry(rq requester) requester {

	if r.ctx.AutoRetries <= 0 {
		return rq
	}

	mw := func(req *http.Request) (*http.Response, error) {
		attempts := r.ctx.AutoRetries
		for {

			// run request
			resp, err := rq(req)
			if err != nil {
				return nil, err
			}

			// check if failed (can be rate limit)
			if resp.StatusCode == http.StatusBadRequest {

				// limit attempts
				attempts--
				if attempts <= 0 {
					err := fmt.Errorf("too many failed attempts for [%s]", req.URL.String())
					data, bodyErr := io.ReadAll(resp.Body)
					if bodyErr == nil {
						err = fmt.Errorf("%s: %s", err, string(data))
					}
					return nil, err
				}

				// wait and retry
				time.Sleep(time.Duration(1) * time.Second)
				continue

			}

			// done
			return resp, nil
		}
	}

	return mw

}

func (r RequestRunner) directConsume(rq requester) (*http.Response, []byte, error) {

	// read response
	res, err := rq(r.ctx.Req)
	if err != nil {
		return nil, nil, err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	return res, data, nil
}

func (r RequestRunner) pagedConsume(rq requester) (*http.Response, []byte, error) {

	// no paging
	if !r.ctx.ConsumeAllPages {
		return r.directConsume(rq)
	}

	// start consuming at first page
	page := 1
	var res *http.Response
	var combinedData bytes.Buffer
	for {

		// set page param
		values := r.ctx.Req.URL.Query()
		values.Set("page", fmt.Sprintf("%d", page))
		r.ctx.Req.URL.RawQuery = values.Encode()

		// read response
		pageRes, err := rq(r.ctx.Req)
		body, err := io.ReadAll(pageRes.Body)
		if err != nil {
			return nil, nil, err
		}
		if err := pageRes.Body.Close(); err != nil {
			return nil, nil, err
		}
		combinedData.Write(body)
		res = pageRes

		// handle paging
		totalPagesRaw := pageRes.Header.Get("X-TOTAL-PAGES")
		if totalPagesRaw == "" {
			break
		}
		totalPages, err := strconv.Atoi(totalPagesRaw)
		if err != nil {
			return nil, nil, err
		}
		if page >= totalPages {
			break
		}
		page++
	}

	return res, combinedData.Bytes(), nil
}
