package surf

import "net/http"

const Version = "0.0.1"

const (
	UserAgent                = "surf/" + Version + " (https://github.com/fupengl/surf)"
	defaultAcceptEncoding    = "gzip, deflate, br"
	defaultAccept            = "application/json, text/plain, */*"
	defaultJsonContentType   = "application/json; charset=UTF-8"
	defaultTextContentType   = "text/plain; charset=UTF-8"
	defaultStreamContentType = "application/octet-stream"
	defaultFormContentType   = "application/x-www-form-urlencoded; charset=UTF-8"
)

var (
	headerUserAgent       = http.CanonicalHeaderKey("User-Agent")
	headerAcceptEncoding  = http.CanonicalHeaderKey("Accept-Encoding")
	headerAccept          = http.CanonicalHeaderKey("Accept")
	headerLocation        = http.CanonicalHeaderKey("Location")
	headerContentEncoding = http.CanonicalHeaderKey("Content-Encoding")
	headerContentType     = http.CanonicalHeaderKey("Content-Type")
	headerContentLength   = http.CanonicalHeaderKey("Content-Length")
)
