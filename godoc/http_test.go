// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godoc

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestServe(t *testing.T) {
	defer func(c http.Client) {
		client = c
	}(client)

	tests := []Response{
		{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       []byte("ok"),
			header:     http.Header{"foo": []string{"bar"}},
		},
		{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       bytes.Repeat([]byte("ok"), 10<<20),
			header:     http.Header{"foo": []string{"bar"}},
		},
		{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       bytes.Repeat([]byte("ok"), 10<<20),
			header:     http.Header{},
		},
	}

	for _, tc := range tests {
		mux := http.NewServeMux()
		mux.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
			if len(tc.header) > 0 {
				copyHeader(w.Header(), tc.header)
			}
			if tc.StatusCode != http.StatusOK {
				http.Error(w, "error", tc.StatusCode)
				return
			}
			data := tc.Body
			for len(data) > 0 {
				if len(data) > 128 {
					w.Write(data[:128])
					data = data[128:]
				} else {
					w.Write(data)
					data = nil
				}
			}
		})
		client = http.Client{Transport: transport{mux}}
		resp, err := Serve("/test")
		if err != nil {
			t.Errorf("Serve returned an error: %v", err)
			continue
		}
		if !reflect.DeepEqual(resp, &tc) {
			t.Errorf("Serve = %s; want %s", resp, &tc)
		}
	}
}

// copyHeader is a narrow copy of http.Header
func copyHeader(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for k, v := range src {
		dst[k] = v
	}
}

func (r *Response) String() string {
	var body string
	if len(r.Body) <= 16 {
		body = string(r.Body)
	} else {
		body = fmt.Sprintf("%s... (len:%d)", r.Body[:16], len(r.Body))
	}
	return fmt.Sprintf("{ Status: %s StatusCode: %d Header: %+v Body: %q }",
		r.Status, r.StatusCode, r.header, body)
}
