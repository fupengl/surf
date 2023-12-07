package surf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	RequestInterceptor  func(config *RequestConfig) (err error)
	ResponseInterceptor func(resp *Response) (err error)

	WithRequestConfig func(c *RequestConfig)

	QuerySerializer struct {
		Encode func(values url.Values) string
	}

	Interceptors struct {
		RequestInterceptors  []RequestInterceptor
		ResponseInterceptors []ResponseInterceptor
	}

	Config struct {
		BaseURL   string
		Headers   http.Header
		Timeout   time.Duration
		Cookies   []*http.Cookie
		CookieJar *http.CookieJar

		Params map[string]string
		Query  url.Values

		QuerySerializer *QuerySerializer

		Interceptors Interceptors

		MaxBodyLength int
		MaxRedirects  int

		Client *http.Client
	}

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

		Interceptors Interceptors

		Body interface{}

		MaxBodyLength int
		MaxRedirects  int

		Client  *http.Client
		Request *http.Request
	}
)

var DefaultConfig = &Config{}

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

func (rc *RequestConfig) SetQuery(key, value string) *RequestConfig {
	if rc.Query == nil {
		rc.Query = make(url.Values)
	}
	rc.Query.Set(key, value)
	return rc
}

func (rc *RequestConfig) SetParams(key, value string) *RequestConfig {
	if rc.Params == nil {
		rc.Params = make(map[string]string)
	}
	rc.Params[key] = value
	return rc
}

func (rc *RequestConfig) SetHeader(key, value string) *RequestConfig {
	if rc.Headers == nil {
		rc.Headers = make(http.Header)
	}
	rc.Headers.Set(key, value)
	return rc
}

func (rc *RequestConfig) SetCookie(cookie *http.Cookie) *RequestConfig {
	rc.Cookies = append(rc.Cookies, cookie)
	return rc
}

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
		rc.Client = &http.Client{Timeout: rc.Timeout}
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
	return rc
}

func (c *Interceptors) AppendRequestInterceptors(interceptors ...RequestInterceptor) *Interceptors {
	if c.RequestInterceptors == nil {
		c.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	c.RequestInterceptors = append(c.RequestInterceptors, interceptors...)
	return c
}

func (c *Interceptors) PrependRequestInterceptors(interceptors ...RequestInterceptor) *Interceptors {
	if c.RequestInterceptors == nil {
		c.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	c.RequestInterceptors = append(interceptors, c.RequestInterceptors...)
	return c
}

func (c *Interceptors) invokeRequestInterceptors(config *RequestConfig) (err error) {
	for _, fn := range c.RequestInterceptors {
		err = fn(config)
		if err != nil {
			return
		}
	}
	return
}

func (c *Interceptors) AppendResponseInterceptors(interceptors ...ResponseInterceptor) *Interceptors {
	if c.ResponseInterceptors == nil {
		c.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	c.ResponseInterceptors = append(c.ResponseInterceptors, interceptors...)
	return c
}

func (c *Interceptors) PrependResponseInterceptors(interceptors ...ResponseInterceptor) *Interceptors {
	if c.ResponseInterceptors == nil {
		c.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	c.ResponseInterceptors = append(interceptors, c.ResponseInterceptors...)
	return c
}

func (c *Interceptors) invokeResponseInterceptors(resp *Response) (err error) {
	for _, fn := range c.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			return
		}
	}
	return
}

func WithBody(body interface{}) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Body = body
	}
}

func WithHeaders(header http.Header) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Headers = header
	}
}

func WithQuery(values url.Values) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Query = values
	}
}

func WithParams(params map[string]string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Params = params
	}
}

func WithCookies(cookies []*http.Cookie) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Cookies = cookies
	}
}

func WithContext(ctx context.Context) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Context = ctx
	}
}

func WithTimeoutContext(ctx context.Context, timeout time.Duration) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Context = ctx
		c.Timeout = timeout
	}
}

func WithSetQuery(key, value string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetQuery(key, value)
	}
}

func WithSetParam(key, value string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetParams(key, value)
	}
}

func WithSetHeader(headers http.Header) WithRequestConfig {
	return func(c *RequestConfig) {
		for k, l := range headers {
			for _, v := range l {
				c.SetHeader(k, v)
			}
		}
	}
}

func WithSetCookie(cookie *http.Cookie) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetCookie(cookie)
	}
}

func combineRequestConfig(args ...WithRequestConfig) RequestConfig {
	config := RequestConfig{}
	for _, arg := range args {
		arg(&config)
	}
	return config
}
