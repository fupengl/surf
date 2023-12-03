package surf

import "errors"

var (
	ErrRequestDataTypeInvalid  = errors.New("request data type is not supported")
	ErrRedirectMissingLocation = errors.New("redirect missing location header")
)
