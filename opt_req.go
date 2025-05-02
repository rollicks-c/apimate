package apimate

import (
	"bytes"
	"encoding/json"
	"github.com/rollicks-c/apimate/internal/client"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func WithCookie(name string, value string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		cookie := &http.Cookie{
			Name:  name,
			Value: value,
		}
		ctx.Req.AddCookie(cookie)
		return nil
	}
}

func WithHeaders(Headers http.Header) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		for key, values := range Headers {
			for _, value := range values {
				ctx.Req.Header.Add(key, value)
			}
		}
		return nil
	}
}
func WithHeader(key, value string) client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Req.Header.Add(key, value)
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

func WithFilePayload(name string, data []byte) client.RequestOption {
	return func(ctx *client.RequestContext) error {

		// write file to form field
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		formFile, err := writer.CreateFormFile("file", name)
		if err != nil {
			return err
		}
		if _, err = io.Copy(formFile, bytes.NewReader(data)); err != nil {
			return err
		}
		_ = writer.Close()

		// create request
		req, err := http.NewRequest(ctx.Method, ctx.Endpoint, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		ctx.Req = req

		return nil
	}
}
