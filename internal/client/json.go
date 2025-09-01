package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
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

type JsonBool bool

func (sb *JsonBool) UnmarshalJSON(data []byte) error {

	// try boolean
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*sb = JsonBool(b)
		return nil
	}

	// try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseBool(strings.ToLower(s))
		if err != nil {
			return fmt.Errorf("invalid boolean string: %q", s)
		}
		*sb = JsonBool(val)
		return nil
	}

	return fmt.Errorf("invalid value for JsonBool: %s", string(data))
}

func (sb *JsonBool) Bool() bool {
	return bool(*sb)
}

type JsonInt64 int64

func (si *JsonInt64) UnmarshalJSON(data []byte) error {
	// try int64
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*si = JsonInt64(i)
		return nil
	}

	// try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 string: %q", s)
		}
		*si = JsonInt64(val)
		return nil
	}

	return fmt.Errorf("invalid value for JsonInt64: %s", string(data))
}

func (si *JsonInt64) Int64() int64 {
	return int64(*si)
}
