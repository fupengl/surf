package surf

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Surf struct {
	Config *Config
	Debug  bool
}

var DefaultClient = &Surf{Config: DefaultConfig}

func New(config *Config) *Surf {
	if config == nil {
		config = DefaultConfig
	}
	return &Surf{
		Config: config,
		Debug:  true,
	}
}

func (s *Surf) prepareRequest(config *RequestConfig) (*http.Request, error) {
	r, err := config.getRequestBody()
	if err != nil {
		return nil, err
	}

	// Auto set Content-type header
	config.setContentTypeHeader()

	req, err := http.NewRequestWithContext(config.Context, config.Method, config.BuildURL(), r)
	if err != nil {
		return nil, err
	}

	// Expose http.Request
	config.Request = req

	// Update global Headers Cookies
	for key, values := range s.Config.Headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}
	for _, cookie := range s.Config.Cookies {
		req.AddCookie(&cookie)
	}

	err = s.Config.invokeRequestInterceptors(config)
	if err != nil {
		return nil, err
	}

	// Update URL
	req.URL, err = url.Parse(config.BuildURL())
	if err != nil {
		return nil, err
	}

	// Update Headers
	for key, values := range config.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Update Cookies
	for _, cookie := range config.Cookies {
		req.AddCookie(&cookie)
	}

	if req.UserAgent() == "" {
		req.Header.Set(headerUserAgent, UserAgent)
	}
	if req.Header.Get(headerAcceptEncoding) == "" {
		req.Header.Set(headerAcceptEncoding, defaultAcceptEncoding)
	}
	if req.Header.Get(headerAccept) == "" {
		req.Header.Set(headerAccept, acceptEncoding)
	}

	if s.Debug {
		log.Printf("DEBUG: Sending request to %s\n", req.URL)
		log.Printf("DEBUG: Request headers: %v\n", req.Header)
		log.Printf("DEBUG: Request cookies: %v\n", req.Cookies())
	}

	return req, nil
}

func (s *Surf) Request(config *RequestConfig) (*Response, error) {
	config.mergeConfig(s.Config)

	req, err := s.prepareRequest(config)
	if err != nil {
		return nil, err
	}

	redirects := 0

	for {
		startTime := time.Now()
		performance := newPerformance()

		resp, err := config.Client.Do(req)
		if err != nil {
			return nil, err
		}

		performance.recordResponseTime(startTime)

		if s.Debug {
			log.Printf("DEBUG: Received response with status code %d\n", resp.StatusCode)
			log.Printf("DEBUG: Response headers: %v\n", resp.Header)
			log.Printf("DEBUG: Response cookies: %v\n", resp.Cookies())
			log.Printf("DEBUG: Response cost: %s\n", performance.ResponseTime)
		}

		if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
			location := resp.Header.Get("Location")
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

		return &response, nil
	}
}

func (s *Surf) makeRequest(url string, method string, args ...WithRequestConfig) (*Response, error) {
	config := combineRequestConfig(args...)
	config.Method = method
	config.Url = url
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

func (s *Surf) CloneDefaultConfig() *Config {
	return &Config{
		BaseURL:              s.Config.BaseURL,
		Headers:              s.Config.Headers.Clone(),
		Timeout:              s.Config.Timeout,
		Cookies:              append([]http.Cookie(nil), s.Config.Cookies...),
		CookieJar:            s.Config.CookieJar,
		QuerySerializer:      s.Config.QuerySerializer,
		RequestInterceptors:  append([]RequestInterceptor(nil), s.Config.RequestInterceptors...),
		ResponseInterceptors: append([]ResponseInterceptor(nil), s.Config.ResponseInterceptors...),
		MaxBodyLength:        s.Config.MaxBodyLength,
		MaxRedirects:         s.Config.MaxRedirects,
		Client:               s.Config.Client,
	}
}
