package surf

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Surf represents the main Surf client configuration.
type Surf struct {
	Config *Config
	Debug  bool
}

// Default is the default Surf instance with the default configuration.
var Default = &Surf{Config: DefaultConfig}

// New creates a new Surf instance with the given configuration.
func New(config *Config) *Surf {
	if config == nil {
		config = DefaultConfig
	}
	if config.Client == nil {
		config.Client = http.DefaultClient
	}
	return &Surf{
		Config: config,
	}
}

// prepareRequest prepares an HTTP request based on the provided configuration.
func (s *Surf) prepareRequest(config *RequestConfig) (*http.Request, error) {
	body, err := config.getRequestBody()
	if err != nil {
		return nil, err
	}

	orgBody := config.Body

	req, err := http.NewRequestWithContext(config.Context, config.Method, config.BuildURL(), body)
	if err != nil {
		return nil, err
	}

	// Expose http.Request
	config.Request = req

	// Update global Headers Cookies
	for key, values := range s.Config.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	for _, cookie := range s.Config.Cookies {
		req.AddCookie(cookie)
	}

	err = s.Config.invokeRequestInterceptors(config)
	if err != nil {
		return nil, err
	}

	err = config.invokeRequestInterceptors(config)
	if err != nil {
		return nil, err
	}

	// Update Request Body
	if orgBody != config.Body {
		newBody, err := config.getRequestBody()
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(newBody)
	}

	// Update Request URL
	req.URL, err = url.Parse(config.BuildURL())
	if err != nil {
		return nil, err
	}

	// Update Request Headers
	for key, values := range config.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Update Request Cookies
	for _, cookie := range config.Cookies {
		req.AddCookie(cookie)
	}

	// Auto set Content-type header
	config.setContentTypeHeader()

	if req.UserAgent() == "" {
		req.Header.Set(headerUserAgent, UserAgent)
	}
	if req.Header.Get(headerAcceptEncoding) == "" {
		req.Header.Set(headerAcceptEncoding, defaultAcceptEncoding)
	}
	if req.Header.Get(headerAccept) == "" {
		req.Header.Set(headerAccept, defaultAccept)
	}

	if s.Debug {
		log.Printf("DEBUG: Sending request to %s\n", req.URL)
		log.Printf("DEBUG: Request headers:\n")
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Printf("	%s: %s\n", key, value)
			}
		}
		log.Printf("DEBUG: Request cookies: %v\n", req.Cookies())
	}

	return req, nil
}

// Request performs an HTTP request using the provided configuration.
func (s *Surf) Request(config *RequestConfig) (*Response, error) {
	config.mergeConfig(s.Config)

	req, err := s.prepareRequest(config)
	if err != nil {
		return nil, err
	}

	redirects := 0

	for {
		performance := &Performance{
			clientTrace: config.clientTrace,
		}

		resp, err := config.Client.Do(req)
		performance.record()
		if err != nil {
			return nil, err
		}

		if s.Debug {
			log.Printf("DEBUG: Received response with status code %d\n", resp.StatusCode)
			log.Printf("DEBUG: Response headers:\n")
			for key, values := range resp.Header {
				for _, value := range values {
					fmt.Printf("	%s: %s\n", key, value)
				}
			}
			log.Printf("DEBUG: Response cookies: %v\n", resp.Cookies())
			log.Printf("DEBUG: Response cost: %s\n", performance.ResponseTime)
		}

		if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
			location := resp.Header.Get(headerLocation)
			if location == "" {
				return nil, ErrRedirectMissingLocation
			}

			originHeader := req.Header.Clone()
			originCookies := req.Cookies()

			// New Request
			req, err = http.NewRequestWithContext(config.Context, config.Method, location, req.Body)
			if err != nil {
				return nil, err
			}

			// Copy headers cookies
			req.Header = originHeader
			for _, cookie := range originCookies {
				req.AddCookie(cookie)
			}

			redirects++
			if config.MaxRedirects > 0 && redirects > config.MaxRedirects {
				return nil, fmt.Errorf("maximum number of redirects (%d) exceeded", config.MaxRedirects)
			}

			continue
		}

		body, err := readBody(resp, config.MaxBodyLength)
		if err != nil {
			return nil, err
		}

		response := Response{
			originalResponse: resp,
			config:           config,
			body:             body,
			Performance:      performance,
		}

		err = s.Config.invokeResponseInterceptors(&response)
		if err != nil {
			return nil, err
		}

		err = config.invokeResponseInterceptors(&response)
		if err != nil {
			return nil, err
		}

		return &response, nil
	}
}

// Upload performs a file upload using the provided URL, file, and optional request configuration.
func (s *Surf) Upload(url string, file *multipartFile, args ...WithRequestConfig) (resp *Response, err error) {
	return s.makeRequest(
		url,
		http.MethodPost,
		append(WithRequestConfigChain{WithBody(file)}, args...)...,
	)
}

// makeRequest is a helper function for creating an HTTP request with default or specified configuration.
func (s *Surf) makeRequest(defaultUrl string, defaultMethod string, args ...WithRequestConfig) (*Response, error) {
	config := combineRequestConfig(args...)
	if config.Url == "" {
		config.Url = defaultUrl
	}
	if config.Method == "" {
		config.Method = defaultMethod
	}
	return s.Request(&config)
}

func (s *Surf) Get(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodGet, args...)
}

func (s *Surf) Post(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodPost, args...)
}

func (s *Surf) Head(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodHead, args...)
}

func (s *Surf) Put(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodPut, args...)
}

func (s *Surf) Patch(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodPatch, args...)
}

func (s *Surf) Delete(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodDelete, args...)
}

func (s *Surf) Options(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodOptions, args...)
}

func (s *Surf) Connect(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodConnect, args...)
}

func (s *Surf) Trace(url string, args ...WithRequestConfig) (*Response, error) {
	return s.makeRequest(url, http.MethodTrace, args...)
}

// CloneDefaultConfig creates a deep copy of the default configuration.
func (s *Surf) CloneDefaultConfig() *Config {
	return &Config{
		BaseURL:              s.Config.BaseURL,
		Header:               s.Config.Header.Clone(),
		Timeout:              s.Config.Timeout,
		Params:               cloneMap(s.Config.Params),
		Query:                cloneURLValues(s.Config.Query),
		Cookies:              append([]*http.Cookie(nil), s.Config.Cookies...),
		CookieJar:            s.Config.CookieJar,
		QuerySerializer:      s.Config.QuerySerializer,
		RequestInterceptors:  append([]RequestInterceptor(nil), s.Config.RequestInterceptors...),
		ResponseInterceptors: append([]ResponseInterceptor(nil), s.Config.ResponseInterceptors...),
		MaxBodyLength:        s.Config.MaxBodyLength,
		MaxRedirects:         s.Config.MaxRedirects,
		Client:               s.Config.Client,
		JSONMarshal:          s.Config.JSONMarshal,
		JSONUnmarshal:        s.Config.JSONUnmarshal,
		XMLMarshal:           s.Config.XMLMarshal,
		XMLUnmarshal:         s.Config.XMLUnmarshal,
	}
}
