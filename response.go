package surf

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Response struct {
	originalResponse *http.Response
	config           *RequestConfig
	body             []byte
	Performance      *Performance
}

func (r *Response) Body() []byte {
	return r.body
}

func (r *Response) Json(v any) error {
	return json.Unmarshal(r.body, &v)
}

func (r *Response) Text() string {
	return string(r.body)
}

func (r *Response) StatusText() string {
	status := r.originalResponse.Status
	parts := strings.SplitN(r.originalResponse.Status, " ", 2)

	if len(parts) < 2 {
		return status
	}

	return parts[1]
}

func (r *Response) Status() int {
	return r.originalResponse.StatusCode
}

func (r *Response) Headers() http.Header {
	return r.originalResponse.Header
}

func (r *Response) Cookies() []*http.Cookie {
	return r.originalResponse.Cookies()
}

func (r *Response) Ok() bool {
	return r.originalResponse.StatusCode >= http.StatusOK && r.originalResponse.StatusCode < http.StatusMultipleChoices
}

func (r *Response) Config() *RequestConfig {
	return r.config
}

func (r *Response) Request() *http.Request {
	return r.originalResponse.Request
}

func (r *Response) OriginalResponse() *http.Response {
	return r.originalResponse
}
