package surf

import (
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
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func readBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()

	size := 0
	contentLength := res.Header.Get("Content-Length")
	if contentLength != "" {
		size, _ = strconv.Atoi(contentLength)
	}
	data, err := readAllInitCap(res.Body, size)
	if err != nil {
		return nil, err
	}

	return data, nil
}
