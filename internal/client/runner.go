package client

import (
	"crypto/tls"
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

	// apply defaults
	for _, opt := range r.ctx.DefaultOptions {
		if err := opt(&r.ctx); err != nil {
			return err
		}
	}

	// prepare
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: r.ctx.SkipTLSVerify},
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
		var reqErr error
		for {

			// run request
			resp, err := rq(req)
			if err != nil {
				runnerErr := fmt.Errorf("failed to execute request [%s] - last error: %v", req.URL.String(), reqErr)
				return nil, runnerErr
			}

			// check if failed (can be rate limit)
			if resp.StatusCode == http.StatusBadRequest {

				// gather error
				data, bodyErr := io.ReadAll(resp.Body)
				if bodyErr == nil {
					reqErr = fmt.Errorf("%s", string(data))
				}

				// limit attempts
				attempts--
				if attempts <= 0 {
					runnerErr := fmt.Errorf("too many failed attempts for [%s] - last error: %v", req.URL.String(), reqErr)
					return nil, runnerErr
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

func (r RequestRunner) directConsume(rq requester) (*http.Response, [][]byte, error) {

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

	//pack
	pack := [][]byte{data}
	return res, pack, nil
}

func (r RequestRunner) pagedConsume(rq requester) (*http.Response, [][]byte, error) {

	// no paging
	if !r.ctx.Paging.ConsumeAll {
		return r.directConsume(rq)
	}

	// start consuming at first page
	page := 1
	var res *http.Response
	var combinedData [][]byte
	for {

		// set page param
		values := r.ctx.Req.URL.Query()
		values.Set(r.ctx.Paging.PageParam, fmt.Sprintf("%d", page))
		r.ctx.Req.URL.RawQuery = values.Encode()

		// read response
		pageRes, err := rq(r.ctx.Req)
		if err != nil {
			return nil, nil, err
		}
		body, err := io.ReadAll(pageRes.Body)
		if err != nil {
			return nil, nil, err
		}
		if err := pageRes.Body.Close(); err != nil {
			return nil, nil, err
		}

		// combine
		combinedData = append(combinedData, body)
		res = pageRes

		// handle paging
		totalPages, ok, err := r.getPageCount(pageRes)
		if !ok {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		if page >= totalPages {
			break
		}
		page++
	}

	return res, combinedData, nil
}

func (r RequestRunner) getPageCount(res *http.Response) (int, bool, error) {

	exp := res.Header.Get(r.ctx.Paging.PageCountHeader)
	if exp == "" {
		return 0, false, nil
	}
	pageCount, err := strconv.Atoi(exp)
	if err != nil {
		return 0, false, err
	}
	return pageCount, true, nil
}
