package apimate

import (
	"bytes"
	"encoding/json"
	"github.com/rollicks-c/apimate/internal/client"
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

func WithNullReceiver() client.RequestOption {
	return func(ctx *client.RequestContext) error {
		ctx.Receiver = func(bytes [][]byte) error {
			return nil
		}
		return nil
	}
}
