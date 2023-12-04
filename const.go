package surf

import "net/http"

const Version = "0.0.1"

const (
	UserAgent             = "surf/" + Version + " (https://github.com/fupengl/surf)"
	defaultAcceptEncoding = "gzip, deflate, br"
	defaultAccept         = "application/json, text/plain, */*'"
)

var (
	headerUserAgent       = http.CanonicalHeaderKey("User-Agent")
	headerAcceptEncoding  = http.CanonicalHeaderKey("Accept-Encoding")
	headerAccept          = http.CanonicalHeaderKey("Accept")
	headerLocation        = http.CanonicalHeaderKey("Location")
	headerContentEncoding = http.CanonicalHeaderKey("Content-Encoding")
	headerContentLength   = http.CanonicalHeaderKey("Content-Length")
)
