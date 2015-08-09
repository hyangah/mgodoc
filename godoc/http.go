// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godoc

// This file contains all machinery to serve contents from a local server.

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var client = http.Client{
	Transport: transport{http.DefaultServeMux},
}

// Serve returns the contents of the url served from a local HTTP server.
func Serve(url string) (*Response, error) {
	initOnce()

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
		header:     resp.Header,
	}, nil
}

// Response wraps http response.
//
// TODO: change Body to io.Reader once gobind supports it.
type Response struct {
	Status     string
	StatusCode int
	Body       []byte

	header http.Header
}

func (r *Response) Header(key string) string {
	return r.header.Get(key)
}

// transport redirects http.Requests to a local default http server.
type transport struct {
	mux *http.ServeMux
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	rw, resc := newResponseWriter()
	go func() {
		t.mux.ServeHTTP(rw, req)
		rw.finish()
	}()
	return <-resc, nil
}

func newResponseWriter() (*populateResponse, <-chan *http.Response) {
	pr, pw := io.Pipe()
	rw := &populateResponse{
		ch: make(chan *http.Response),
		pw: pw,
		res: &http.Response{
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			Header:     make(http.Header),
			Close:      true,
			Body:       pr,
		},
	}
	return rw, rw.ch
}

// populateResponse is a http.ResponseWriter that populates the Response
// in res, and writes its body to a pipe connected to the response body.
// Once writes begin or finish is called, the response is sent on ch.
type populateResponse struct {
	res          *http.Response
	ch           chan *http.Response
	wroteHeader  bool
	hasContent   bool
	sentResponse bool
	pw           *io.PipeWriter
}

func (pr *populateResponse) finish() {
	if !pr.wroteHeader {
		pr.WriteHeader(500)
	}
	if !pr.sentResponse {
		pr.sendResponse()
	}
	pr.pw.Close()
}

func (pr *populateResponse) sendResponse() {
	if pr.sentResponse {
		return
	}
	pr.sentResponse = true

	if pr.hasContent {
		pr.res.ContentLength = -1
	}
	pr.ch <- pr.res
}

func (pr *populateResponse) Header() http.Header {
	return pr.res.Header
}

func (pr *populateResponse) WriteHeader(code int) {
	if pr.wroteHeader {
		return
	}
	pr.wroteHeader = true

	pr.res.StatusCode = code
	pr.res.Status = fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func (pr *populateResponse) Write(p []byte) (n int, err error) {
	if !pr.wroteHeader {
		pr.WriteHeader(http.StatusOK)
	}
	pr.hasContent = true
	if !pr.sentResponse {
		pr.sendResponse()
	}
	return pr.pw.Write(p)
}
