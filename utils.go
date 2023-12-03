package surf

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

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
		}
		return b, err
	}
}

func readBody(res *http.Response, maxBodyLength int) ([]byte, error) {
	defer res.Body.Close()

	size := 0
	contentLength := res.Header.Get("Content-Length")
	if contentLength != "" {
		size, _ = strconv.Atoi(contentLength)
	}

	if maxBodyLength > 0 && size > maxBodyLength {
		return nil, fmt.Errorf("response body exceeds the maximum length of %d", maxBodyLength)
	}

	data, err := readAllInitCap(res.Body, size)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
