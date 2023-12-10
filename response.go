package surf

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

// Response represents the HTTP response received after sending a request.
type Response struct {
	originalResponse *http.Response
	config           *RequestConfig
	body             []byte
	Performance      *Performance
}

// Body returns the raw body of the HTTP response.
func (r *Response) Body() []byte {
	return r.body
}

// BodyReader returns the response body as an io.Reader.
func (r *Response) BodyReader() io.Reader {
	return bytes.NewReader(r.body)
}

// Json parses the JSON response body and stores the result in the provided variable (v).
func (r *Response) Json(v interface{}) error {
	return json.Unmarshal(r.body, &v)
}

// Text returns the response body as a string.
func (r *Response) Text() string {
	return string(r.body)
}

// SaveToFile saves the response body to a file with the specified filename.
func (r *Response) SaveToFile(filename string) error {
	err := os.WriteFile(filename, r.body, 0644)
	return err
}

// StatusText returns the status text part of the HTTP status code and reason.
func (r *Response) StatusText() string {
	status := r.originalResponse.Status
	parts := strings.SplitN(r.originalResponse.Status, " ", 2)

	if len(parts) < 2 {
		return status
	}

	return parts[1]
}

// Status returns the HTTP status code of the response.
func (r *Response) Status() int {
	return r.originalResponse.StatusCode
}

// Headers returns the HTTP headers of the response.
func (r *Response) Headers() http.Header {
	return r.originalResponse.Header
}

// Cookies returns the cookies set in the HTTP response.
func (r *Response) Cookies() []*http.Cookie {
	return r.originalResponse.Cookies()
}

// Ok checks if the HTTP response status code indicates success (2xx).
func (r *Response) Ok() bool {
	return r.originalResponse.StatusCode >= http.StatusOK && r.originalResponse.StatusCode < http.StatusMultipleChoices
}

// Config returns the request configuration associated with the response.
func (r *Response) Config() *RequestConfig {
	return r.config
}

// ContentEncoding returns the content encoding specified in the response header.
// It retrieves the value of the "Content-Encoding" header, indicating the encoding
// transformation that has been applied to the response body, such as "gzip" or "deflate".
// If the header is not present, an empty string is returned.
func (r *Response) ContentEncoding() string {
	return r.originalResponse.Header.Get(headerContentEncoding)
}

// Request returns the original HTTP request associated with the response.
func (r *Response) Request() *http.Request {
	return r.originalResponse.Request
}

// OriginalResponse returns the original HTTP response.
func (r *Response) OriginalResponse() *http.Response {
	return r.originalResponse
}
