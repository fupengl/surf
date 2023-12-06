package surf

import (
	"log"
	"net/http"
	"net/url"
	"testing"
)

type ComparativeData struct {
	RequestConfig
	Output string
}

func TestRequestConfig_BuildURL(t *testing.T) {
	data := [...]ComparativeData{
		{
			RequestConfig: RequestConfig{
				BaseURL: "https://github.com",
				Url:     "/fupengl",
			},
			Output: "https://github.com/fupengl",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "https://github.com",
				Url:     "fupengl",
			},
			Output: "https://github.com/fupengl",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "https://github.com/",
				Url:     "/fupengl",
			},
			Output: "https://github.com/fupengl",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "",
				Url:     "/fupengl",
			},
			Output: "/fupengl",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "",
				Url:     "https://www.baidu.com",
			},
			Output: "https://www.baidu.com",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "www.baidu.com",
				Url:     "a",
			},
			Output: "https://www.baidu.com/a",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "www.baidu.com",
				Url:     "a",
				Query: url.Values{
					"a": {"a"},
				},
			},
			Output: "https://www.baidu.com/a?a=a",
		},
		{
			RequestConfig: RequestConfig{
				BaseURL: "www.baidu.com/:id",
				Url:     "a",
				Query: url.Values{
					"a": {"a"},
				},
				Params: map[string]string{
					"id": "xxx",
				},
			},
			Output: "https://www.baidu.com/xxx/a?a=a",
		},
	}

	for _, item := range data {
		output := item.RequestConfig.BuildURL()

		if output != item.Output {
			t.Fatalf("url build expect %s output %s.", item.Output, output)
		}
	}
}

func TestRequestConfig_BuildQuery(t *testing.T) {
	data := [...]ComparativeData{
		{
			RequestConfig: RequestConfig{
				QuerySerializer: &QuerySerializer{},
				Query: map[string][]string{
					"a": {"1"},
				},
			},
			Output: "a=1",
		},
		{
			RequestConfig: RequestConfig{
				QuerySerializer: &QuerySerializer{
					Encode: func(values url.Values) string {
						return ""
					},
				},
				Query: map[string][]string{
					"a": {"1"},
				},
			},
			Output: "",
		},
	}

	for _, item := range data {
		outPut := item.RequestConfig.BuildQuery()

		if outPut != item.Output {
			t.Fatalf("query build expect %s output %s.", item.Output, outPut)
		}
	}
}

func TestRequestConfig_SetCookie(t *testing.T) {
	config := RequestConfig{
		Cookies: nil,
	}
	config.SetCookie(&http.Cookie{
		Name:  "1",
		Value: "1",
	})
	if len(config.Cookies) != 1 {
		log.Fatal("set cookie error")
	}
}
