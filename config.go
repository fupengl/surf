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

	Config struct {
		BaseURL   string
		Headers   http.Header
		Timeout   time.Duration
		Cookies   []http.Cookie
		CookieJar *http.CookieJar

		QuerySerializer *QuerySerializer

		RequestInterceptors  []RequestInterceptor
		ResponseInterceptors []ResponseInterceptor

		MaxBodyLength int
		MaxRedirects  int

		Client *http.Client
	}

	RequestConfig struct {
		BaseURL string
		Url     string
		Headers http.Header
		Method  string
		Cookies []http.Cookie

		Timeout time.Duration
		Context context.Context

		Query           url.Values
		QuerySerializer *QuerySerializer

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

func (rc *RequestConfig) appendQueryToURL(u string) string {
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

func (rc *RequestConfig) SetQuery(key, value string) *RequestConfig {
	if rc.Query == nil {
		rc.Query = make(url.Values)
	}
	rc.Query.Set(key, value)
	return rc
}

func (rc *RequestConfig) getRequestBody() (r io.Reader, err error) {
	if rc.Body == nil {
		return
	}
	data := rc.Body
	r, ok := data.(io.Reader)
	if ok {
		return r, nil
	}

	buf, ok := data.([]byte)
	if !ok {
		err = ErrRequestDataTypeInvalid
		return
	}
	r = bytes.NewReader(buf)
	return
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
	return rc
}

func (c *Config) AppendRequestInterceptors(interceptors ...RequestInterceptor) *Config {
	if c.RequestInterceptors == nil {
		c.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	c.RequestInterceptors = append(c.RequestInterceptors, interceptors...)
	return c
}

func (c *Config) PrependRequestInterceptors(interceptors ...RequestInterceptor) *Config {
	if c.RequestInterceptors == nil {
		c.RequestInterceptors = make([]RequestInterceptor, 0)
	}
	c.RequestInterceptors = append(interceptors, c.RequestInterceptors...)
	return c
}

func (c *Config) invokeRequestInterceptors(config *RequestConfig) (err error) {
	for _, fn := range c.RequestInterceptors {
		err = fn(config)
		if err != nil {
			return
		}
	}
	return
}

func (c *Config) AppendResponseInterceptors(interceptors ...ResponseInterceptor) *Config {
	if c.ResponseInterceptors == nil {
		c.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	c.ResponseInterceptors = append(c.ResponseInterceptors, interceptors...)
	return c
}

func (c *Config) PrependResponseInterceptors(interceptors ...ResponseInterceptor) *Config {
	if c.ResponseInterceptors == nil {
		c.ResponseInterceptors = make([]ResponseInterceptor, 0)
	}
	c.ResponseInterceptors = append(interceptors, c.ResponseInterceptors...)
	return c
}

func (c *Config) invokeResponseInterceptors(resp *Response) (err error) {
	for _, fn := range c.ResponseInterceptors {
		err = fn(resp)
		if err != nil {
			return
		}
	}
	return
}

func WithBody(values interface{}) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Body = values
	}
}

func WithQuery(values url.Values) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Query = values
	}
}

func WithHeader(values http.Header) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Headers = values
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

func combineRequestConfig(args ...WithRequestConfig) RequestConfig {
	config := RequestConfig{}
	for _, arg := range args {
		arg(&config)
	}
	return config
}
