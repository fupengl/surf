package surf

import (
	"net/http"
	"net/url"
)

type Surf struct {
	Config *Config
}

var DefaultClient = &Surf{DefaultConfig}

func New(config *Config) *Surf {
	if config == nil {
		config = &Config{}
	}
	return &Surf{config}
}

func (s *Surf) prepareRequest(config *RequestConfig) (*http.Request, error) {
	r, err := config.getRequestBody()
	if err != nil {
		return nil, err
	}

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
			req.Header.Set(key, value)
		}
	}

	// Update Cookies
	for _, cookie := range config.Cookies {
		req.AddCookie(&cookie)
	}

	if req.UserAgent() == "" {
		req.Header.Set(headerUserAgent, UserAgent)
	}

	return req, nil
}

func (s *Surf) Request(config *RequestConfig) (*Response, error) {
	config.mergeConfig(s.Config)

	req, err := s.prepareRequest(config)
	if err != nil {
		return nil, err
	}

	resp, err := config.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}

	response := Response{
		originalResponse: resp,
		config:           config,
		body:             body,
	}

	err = s.Config.invokeResponseInterceptors(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
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
