package client

import (
	"bytes"
	"encoding/json"
	"io"
)

func ParseArrayList(raw []byte) ([]byte, error) {

	buf := bytes.NewReader(raw)
	decoder := json.NewDecoder(buf)

	var all []map[string]interface{}

	// Decode JSON arrays one by one
	for {
		var page []map[string]interface{}
		if err := decoder.Decode(&page); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		all = append(all, page...)
	}

	//
	merged, err := json.Marshal(all)
	if err != nil {
		return nil, err
	}
	return merged, nil
}
