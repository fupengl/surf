package surf

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// WithRequestConfig is a function signature for configuring request options.
type WithRequestConfig func(c *RequestConfig)

type WithRequestConfigChain []WithRequestConfig

// WithBody sets the request body in the request configuration.
func WithBody(body interface{}) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Body = body
	}
}

// WithHeaders sets the request headers in the request configuration.
func WithHeaders(header http.Header) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Headers = header
	}
}

// WithQuery sets the query parameters in the request configuration.
func WithQuery(values url.Values) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Query = values
	}
}

// WithParams sets the parameters in the request configuration.
func WithParams(params map[string]string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Params = params
	}
}

// WithCookies sets the cookies in the request configuration.
func WithCookies(cookies []*http.Cookie) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Cookies = cookies
	}
}

// WithContext sets the context in the request configuration.
func WithContext(ctx context.Context) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Context = ctx
	}
}

// WithTimeoutContext sets the context and timeout in the request configuration.
func WithTimeoutContext(ctx context.Context, timeout time.Duration) WithRequestConfig {
	return func(c *RequestConfig) {
		c.Context = ctx
		c.Timeout = timeout
	}
}

// WithSetQuery adds a query parameter in the request configuration.
func WithSetQuery(key, value string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetQuery(key, value)
	}
}

// WithSetParam adds a parameter in the request configuration.
func WithSetParam(key, value string) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetParam(key, value)
	}
}

// WithSetHeader adds a header in the request configuration.
func WithSetHeader(headers http.Header) WithRequestConfig {
	return func(c *RequestConfig) {
		for k, l := range headers {
			for _, v := range l {
				c.SetHeader(k, v)
			}
		}
	}
}

// WithSetCookie adds a cookie in the request configuration.
func WithSetCookie(cookie *http.Cookie) WithRequestConfig {
	return func(c *RequestConfig) {
		c.SetCookie(cookie)
	}
}

// WithRequestInterceptor append RequestInterceptor in the request configuration.
func WithRequestInterceptor(handler RequestInterceptor) WithRequestConfig {
	return func(c *RequestConfig) {
		c.AppendRequestInterceptors(handler)
	}
}

// WithResponseInterceptor append ResponseInterceptor in the request configuration.
func WithResponseInterceptor(handler ResponseInterceptor) WithRequestConfig {
	return func(c *RequestConfig) {
		c.AppendResponseInterceptors(handler)
	}
}

// combineRequestConfig combines multiple request configurations into a single configuration.
func combineRequestConfig(args ...WithRequestConfig) RequestConfig {
	config := RequestConfig{}
	for _, arg := range args {
		arg(&config)
	}
	return config
}
