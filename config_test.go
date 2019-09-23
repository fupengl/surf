package surf

import (
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
