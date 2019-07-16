package lash

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	// HTTPRequest building request
	HTTPRequest struct {
		session  *Session
		serr     SessionErr
		req      *http.Request
		statuses []int
		Client   *http.Client
	}
	// HTTPResponse from a request
	HTTPResponse struct {
		session  *Session
		response *http.Response
		body     []byte
	}
)

// Curl wrapper for simple http client, uses the default session
func Curl(url string) *HTTPRequest {
	s := DefaultSession
	if s == nil {
		s = NewSession()
	}
	return s.Curl(url)
}

// Curl for this session
func (s *Session) Curl(url string) *HTTPRequest {
	serr := SessionErr{Type: "HTTPRequest"}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.err = serr.fail("Curl", err)
	}
	return &HTTPRequest{
		serr:     serr,
		session:  s,
		req:      req,
		statuses: []int{200, 201, 202, 204},
	}
}

// Post method will be used
func (cmd *HTTPRequest) Post() *HTTPRequest {
	cmd.req.Method = http.MethodPost
	return cmd
}

// Put method will be used
func (cmd *HTTPRequest) Put() *HTTPRequest {
	cmd.req.Method = http.MethodPut
	return cmd
}

// Delete method will be used
func (cmd *HTTPRequest) Delete() *HTTPRequest {
	cmd.req.Method = http.MethodDelete
	return cmd
}

// AllowResponses overrides the default happy path (200, 201, 202, 204)
func (cmd *HTTPRequest) AllowResponses(status ...int) *HTTPRequest {
	cmd.statuses = status
	return cmd
}

// Response the request
func (cmd *HTTPRequest) Response() *HTTPResponse {
	r := &HTTPResponse{session: cmd.session}
	if cmd.Client == nil {
		cmd.Client = &http.Client{}
	}
	var err error
	r.response, err = cmd.Client.Do(cmd.req)
	if err != nil {
		cmd.session.err = cmd.serr.fail("Send", err)
		return r
	}

	if !isInList(r.response.StatusCode, cmd.statuses) {
		err = fmt.Errorf("status %v not allowed", r.response.StatusCode)
		cmd.session.err = cmd.serr.fail("Send", err)
		return r
	}

	if r.response.Body != nil {
		defer func() { _ = r.response.Body.Close() }()
		r.body, err = ioutil.ReadAll(r.response.Body)
		if err != nil {
			cmd.session.err = cmd.serr.fail("ReadBody", err)
		}
	}

	return r
}
func isInList(search int, list []int) bool {
	for _, s := range list {
		if s == search {
			return true
		}
	}
	return false
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
	return r.session != nil && r.session.err != nil
}
