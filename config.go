package surf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type (
	// RequestInterceptor defines a function signature for request interceptors.
	RequestInterceptor func(config *RequestConfig) (err error)
	// ResponseInterceptor defines a function signature for response interceptors.
	ResponseInterceptor func(resp *Response) (err error)

	// QuerySerializer is responsible for encoding URL query parameters.
	QuerySerializer struct {
		Encode func(values url.Values) string
	}

	// Config holds the configuration for Surf.
	Config struct {
		BaseURL   string
		Headers   http.Header
		Timeout   time.Duration
		Cookies   []*http.Cookie
		CookieJar *http.CookieJar

		Params map[string]string
		Query  url.Values

		QuerySerializer *QuerySerializer

		RequestInterceptors  []RequestInterceptor
		ResponseInterceptors []ResponseInterceptor

		requestInterceptorsMu  sync.RWMutex
		responseInterceptorsMu sync.RWMutex

		MaxBodyLength int
		MaxRedirects  int

		Client *http.Client
	}

	// RequestConfig holds the configuration for a specific HTTP request.
	RequestConfig struct {
		BaseURL string
		Url     string
		Headers http.Header
		Method  string
		Cookies []*http.Cookie

		Timeout time.Duration
		Context context.Context

		Params map[string]string

		Query           url.Values
		QuerySerializer *QuerySerializer

		RequestInterceptors  []RequestInterceptor
		ResponseInterceptors []ResponseInterceptor

		requestInterceptorsMu  sync.Mutex
		responseInterceptorsMu sync.Mutex

		Body interface{}

		MaxBodyLength int
		MaxRedirects  int

		Client  *http.Client
		Request *http.Request

		clientTrace *clientTrace
	}
)

// DefaultConfig is the default configuration for Surf.
var DefaultConfig = &Config{
	Client: http.DefaultClient,
}

// BuildURL constructs the full URL based on the configuration.
func (rc *RequestConfig) BuildURL() string {
	baseURL := rc.BaseURL

	if baseURL == "" {
		return rc.appendQueryToURL(rc.Url)
	}

	if strings.Contains(rc.Url, "://") {
		return rc.Url
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	urlPath := strings.TrimLeft(rc.Url, "/")

	if !strings.Contains(baseURL, "://") {
		baseURL = "https://" + baseURL
	}

	u, err := url.Parse(baseURL + urlPath)
	if err != nil {
		return ""
	}

	return rc.appendQueryToURL(u.String())
}

// BuildQuery constructs the query string based on the configuration.
func (rc *RequestConfig) BuildQuery() string {
	var qs string
	if rc.Query != nil {
		if rc.QuerySerializer != nil && rc.QuerySerializer.Encode != nil {
			qs = rc.QuerySerializer.Encode(rc.Query)
		} else {
			qs = rc.Query.Encode()
		}
	}
	return qs
}

// SetQuery sets a query parameter in the request configuration.
func (rc *RequestConfig) SetQuery(key, value string) *RequestConfig {
	if rc.Query == nil {
		rc.Query = make(url.Values)
	}
	rc.Query.Set(key, value)
	return rc
}

// SetParams sets a parameter in the request configuration.
func (rc *RequestConfig) SetParams(key, value string) *RequestConfig {
	if rc.Params == nil {
		rc.Params = make(map[string]string)
	}
	rc.Params[key] = value
	return rc
}

// SetHeader sets a header in the request configuration.
func (rc *RequestConfig) SetHeader(key, value string) *RequestConfig {
	if rc.Headers == nil {
		rc.Headers = make(http.Header)
	}
	rc.Headers.Set(key, value)
	return rc
}

// SetCookie adds a cookie to the request configuration.
func (rc *RequestConfig) SetCookie(cookie *http.Cookie) *RequestConfig {
	rc.Cookies = append(rc.Cookies, cookie)
	return rc
}

// appendQueryToURL appends query parameters to the URL in the request configuration.
func (rc *RequestConfig) appendQueryToURL(u string) string {
	if rc.Params != nil {
		for key, value := range rc.Params {
			placeholder := ":" + key
			u = strings.Replace(u, placeholder, value, -1)
		}
	}

	if rc.Query != nil {
		qs := rc.BuildQuery()
		if strings.Contains(u, "?") {
			return u + "&" + qs
		} else {
			return u + "?" + qs
		}
	}
	return u
}

// getRequestBody returns the request body based on the configured body type.
func (rc *RequestConfig) getRequestBody() (r io.Reader, err error) {
	if rc.Body == nil {
		return
	}

	switch data := rc.Body.(type) {
	case io.Reader:
		return data, nil
	case []byte:
		return bytes.NewReader(data), nil
	default:
		err = ErrRequestDataTypeInvalid
		return
	}
}

// setContentTypeHeader sets the Content-Type header based on the request body type.
func (rc *RequestConfig) setContentTypeHeader() {
	if rc.Headers.Get(headerContentType) != "" {
		return
	}

	switch body := rc.Body.(type) {
	case string:
		rc.SetHeader(headerContentType, defaultTextContentType)
	case []byte:
		rc.SetHeader(headerContentType, defaultStreamContentType)
	case io.Reader:
		// Do nothing, assuming the user has set the appropriate Content-Type
	case url.Values:
		// For form data, set Content-Type as application/x-www-form-urlencoded
		rc.SetHeader(headerContentType, defaultFormContentType)
		// URL encode the form data
		rc.Body = strings.NewReader(body.Encode())
	default:
		// For other types, set the default Content-Type as JSON
		rc.SetHeader(headerContentType, defaultJsonContentType)
	}
}

// mergeConfig merges the current request configuration with the Config.
func (rc *RequestConfig) mergeConfig(config *Config) *RequestConfig {
	if rc.BaseURL == "" {
		rc.BaseURL = config.BaseURL
	}

	if rc.Client == nil {
		rc.Client = config.Client
	}

	if rc.Timeout == 0 {
		rc.Timeout = config.Timeout
	}

	if rc.Client == nil {
		rc.Client = http.DefaultClient
	}

	if config.CookieJar != nil {
		rc.Client.Jar = *config.CookieJar
	}

	if rc.Method == "" {
		rc.Method = http.MethodGet
	}

	if rc.QuerySerializer == nil {
		rc.QuerySerializer = config.QuerySerializer
	}

	if rc.Context == nil {
		rc.Context = context.Background()
	}

	if rc.MaxBodyLength == 0 {
		rc.MaxBodyLength = config.MaxBodyLength
	}

	if config.Params != nil {
		for key, val := range config.Params {
			if _, ok := rc.Params[key]; !ok {
				rc.SetParams(key, val)
			}
		}
	}

	if config.Query != nil {
		for key, val := range config.Query {
			if !rc.Query.Has(key) {
				for _, s := range val {
					rc.SetQuery(key, s)
				}
			}
		}
	}

	// Enable http trace for Performance
	rc.clientTrace = &clientTrace{}
	rc.Context = rc.clientTrace.createContext(rc.Context)
	return rc
}

// AppendRequestInterceptors appends request interceptors to the interceptor list.
func (rc *RequestConfig) AppendRequestInterceptors(interceptors ...RequestInterceptor) *RequestConfig {
	rc.requestInterceptorsMu.Lock()
	defer rc.requestInterceptorsMu.Unlock()

	if rc.RequestInterceptors == nil {
		rc.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	rc.RequestInterceptors = append(rc.RequestInterceptors, interceptors...)
	return rc
}

// PrependRequestInterceptors prepends request interceptors to the interceptor list.
func (rc *RequestConfig) PrependRequestInterceptors(interceptors ...RequestInterceptor) *RequestConfig {
	rc.requestInterceptorsMu.Lock()
	defer rc.requestInterceptorsMu.Unlock()

	if rc.RequestInterceptors == nil {
		rc.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	rc.RequestInterceptors = append(interceptors, rc.RequestInterceptors...)
	return rc
}

// invokeRequestInterceptors invokes all request interceptors with the provided configuration.
func (rc *RequestConfig) invokeRequestInterceptors(config *RequestConfig) (err error) {
	rc.requestInterceptorsMu.Lock()
	defer rc.requestInterceptorsMu.Unlock()

	for _, fn := range rc.RequestInterceptors {
		err = fn(config)
		if err != nil {
			return
		}
	}
	return
}

// AppendResponseInterceptors appends response interceptors to the interceptor list.
func (rc *RequestConfig) AppendResponseInterceptors(interceptors ...ResponseInterceptor) *RequestConfig {
	rc.responseInterceptorsMu.Lock()
	defer rc.responseInterceptorsMu.Unlock()

	if rc.ResponseInterceptors == nil {
		rc.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	rc.ResponseInterceptors = append(rc.ResponseInterceptors, interceptors...)
	return rc
}

// PrependResponseInterceptors prepends response interceptors to the interceptor list.
func (rc *RequestConfig) PrependResponseInterceptors(interceptors ...ResponseInterceptor) *RequestConfig {
	rc.responseInterceptorsMu.Lock()
	defer rc.responseInterceptorsMu.Unlock()

	if rc.ResponseInterceptors == nil {
		rc.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	rc.ResponseInterceptors = append(interceptors, rc.ResponseInterceptors...)
	return rc
}

// invokeResponseInterceptors invokes all response interceptors with the provided response.
func (rc *RequestConfig) invokeResponseInterceptors(resp *Response) (err error) {
	rc.responseInterceptorsMu.Lock()
	defer rc.responseInterceptorsMu.Unlock()

	for _, fn := range rc.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			return
		}
	}
	return
}

// invokeRequestInterceptors invokes all request interceptors with the provided configuration.
func (c *Config) invokeRequestInterceptors(config *RequestConfig) (err error) {
	c.requestInterceptorsMu.Lock()
	defer c.requestInterceptorsMu.Unlock()

	for _, fn := range c.RequestInterceptors {
		err = fn(config)
		if err != nil {
			return
		}
	}
	return
}

// invokeResponseInterceptors invokes all response interceptors with the provided response.
func (c *Config) invokeResponseInterceptors(resp *Response) (err error) {
	c.responseInterceptorsMu.Lock()
	defer c.responseInterceptorsMu.Unlock()

	for _, fn := range c.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			return
		}
	}
	return
}
