package surf

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/dsnet/compress/brotli"
)

func readBody(res *http.Response, maxBodyLength int) ([]byte, error) {
	defer res.Body.Close()

	var reader io.Reader = res.Body

	// Check for Content-Encoding and decode accordingly
	// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Content-Encoding
	encoding := res.Header.Get(headerContentEncoding)
	// If no content, but headers still say that it is encoded,
	if res.StatusCode != http.StatusNoContent || res.Request.Method != http.MethodHead {
		var err error
		switch encoding {
		case "gzip", "x-gzip", "compress", "x-compress":
			reader, err = gzip.NewReader(res.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to create Gzip reader: %w", err)
			}
			defer reader.(*gzip.Reader).Close()
		case "br":
			reader, err = brotli.NewReader(res.Body, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create Brotli reader: %w", err)
			}
			defer reader.(*brotli.Reader).Close()
		case "deflate":
			reader = flate.NewReader(res.Body)
			defer reader.(io.ReadCloser).Close()
		}
	}

	size := 0
	contentLength := res.Header.Get(headerContentLength)
	if contentLength != "" {
		size, _ = strconv.Atoi(contentLength)
	}

	if maxBodyLength > 0 && size > maxBodyLength {
		return nil, fmt.Errorf("response body exceeds the maximum length of %d", maxBodyLength)
	}

	data, err := readAllInitCap(reader, size)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func readAllInitCap(r io.Reader, initCap int) ([]byte, error) {
	if initCap <= 0 {
		initCap = 512
	}
	b := make([]byte, 0, initCap)
	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}
	return b, nil
}

func cloneMap[V any](originalMap map[string]V) map[string]V {
	clonedMap := make(map[string]V)

	for key, value := range originalMap {
		clonedMap[key] = value
	}

	return clonedMap
}

func cloneURLValues(originalValues url.Values) url.Values {
	clonedValues := make(url.Values)

	for key, values := range originalValues {
		clonedValues[key] = append([]string(nil), values...)
	}

	return clonedValues
}

func isZero(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func defaultValue[T any](value, defaultValue T) T {
	if isZero(value) {
		return defaultValue
	}
	return value
}
