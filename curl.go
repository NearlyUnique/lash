package lash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	// HTTPRequest building request
	HTTPRequest struct {
		scope    *Scope
		serr     ScopeErr
		Req      *http.Request
		statuses []int
		Client   *http.Client
	}
	// HTTPResponse from a request
	HTTPResponse struct {
		scope    *Scope
		response *http.Response
		body     []byte
	}
)

const AnyHTTPStatus = 9999

// Curl for this scope
func (s *Scope) Curl(url string, args ...interface{}) *HTTPRequest {
	serr := ScopeErr{Type: "HTTPRequest"}
	req, err := http.NewRequest("GET", s.EnvStr(url, args...), nil)
	if err != nil {
		s.SetErr(serr.fail("Curl", err))
	}
	return &HTTPRequest{
		serr:     serr,
		scope:    s,
		Req:      req,
		statuses: []int{200, 201, 202, 204},
	}
}

// Post method will be used
func (cmd *HTTPRequest) Post(body []byte) *HTTPRequest {
	return cmd.Method(http.MethodPost, body)
}

// Put method will be used
func (cmd *HTTPRequest) Put(body []byte) *HTTPRequest {
	return cmd.Method(http.MethodPut, body)
}

// Delete method will be used
func (cmd *HTTPRequest) Delete() *HTTPRequest {
	return cmd.Method(http.MethodDelete, nil)
}

// Method can be any
func (cmd *HTTPRequest) Method(method string, body []byte) *HTTPRequest {
	cmd.Req.Method = method
	if body != nil {
		cmd.Req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return cmd
}

// AllowResponses overrides the default happy path (200, 201, 202, 204)
func (cmd *HTTPRequest) AllowResponses(status ...int) *HTTPRequest {
	cmd.statuses = status
	return cmd
}

// Response the request
func (cmd *HTTPRequest) Response() *HTTPResponse {
	r := &HTTPResponse{scope: cmd.scope}
	if cmd.Client == nil {
		cmd.Client = &http.Client{}
	}
	var err error
	r.response, err = cmd.Client.Do(cmd.Req)
	if err != nil {
		cmd.scope.SetErr(cmd.serr.fail("Send", err))
		return r
	}

	if !isInList(r.response.StatusCode, cmd.statuses) {
		err = fmt.Errorf("status %v not allowed", r.response.StatusCode)
		cmd.scope.SetErr(cmd.serr.fail("Send", err))
		return r
	}

	if r.response.Body != nil {
		defer func() { _ = r.response.Body.Close() }()
		r.body, err = ioutil.ReadAll(r.response.Body)
		if err != nil {
			cmd.scope.SetErr(cmd.serr.fail("ReadBody", err))
		}
	}

	return r
}

// Header can be set, this overwrites and previous value
func (cmd *HTTPRequest) Header(name, value string, args ...interface{}) *HTTPRequest {
	cmd.Req.Header.Set(name, cmd.scope.EnvStr(value, args...))
	return cmd
}

// AddHeader can be set, this allows multiple values for the same header
func (cmd *HTTPRequest) AddHeader(name, value string, args ...interface{}) *HTTPRequest {
	cmd.Req.Header.Add(name, cmd.scope.EnvStr(value, args...))
	return cmd
}

// CommonFunc so allow simplifying common values
func (cmd *HTTPRequest) CommonFunc(custom func(r *HTTPRequest)) *HTTPRequest {
	if custom != nil {
		custom(cmd)
	}
	return cmd
}

// StatusCode of the request, will be zero if no valid status exists
func (r *HTTPResponse) StatusCode() int {
	if r == nil || r.response == nil {
		return 0
	}
	return r.response.StatusCode
}

// BodyString body as a string
func (r *HTTPResponse) BodyString() string {
	if r == nil || r.body == nil {
		return ""
	}
	return string(r.body)
}

// BodyBytes body as a []byte
func (r *HTTPResponse) BodyBytes() []byte {
	if r == nil {
		return nil
	}
	return r.body
}

func (r *HTTPResponse) IsError() bool {
	return r.scope != nil && r.scope.err != nil
}

// BodyJSON puts the response body into the passed in type
// returns false if no body
func (r *HTTPResponse) FromJSON(buf interface{}) bool {
	if r == nil || r.body == nil {
		return false
	}
	if err := json.Unmarshal(r.body, buf); err != nil {
		r.scope.SetErr(&ScopeErr{Type: "HTTPResponse", Action: "FromJSON", Err: err})
		return false
	}

	return true

}

func isInList(search int, list []int) bool {
	for _, s := range list {
		if s == AnyHTTPStatus || s == search {
			return true
		}
	}
	return false
}
