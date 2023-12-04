package surf

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dsnet/compress/brotli"
)

func readBody(res *http.Response, maxBodyLength int) ([]byte, error) {
	defer res.Body.Close()

	var reader io.Reader = res.Body

	// Check for Content-Encoding and decode accordingly
	// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Content-Encoding
	encoding := res.Header.Get(headerContentEncoding)
	switch encoding {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gzip reader: %w", err)
		}
		defer reader.(*gzip.Reader).Close()
	case "br":
		var err error
		reader, err = brotli.NewReader(res.Body, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Brotli reader: %w", err)
		}
	case "deflate":
		reader = flate.NewReader(res.Body)
		defer reader.(io.ReadCloser).Close()
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
