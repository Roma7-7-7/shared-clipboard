package domain

import "fmt"

type (
	ErrorCode string
)

var (
	ErrorBadRequest              = ErrorCode("ERR_0400")
	ErrorCodeNotFound            = ErrorCode("ERR_0404")
	ErrorCodeMethodNotAllowed    = ErrorCode("ERR_0405")
	ErrorCodeInternalServerError = ErrorCode("ERR_0500")

	ErrorCodeMarshalResponse = ErrorCode("ERR_1001")

	ErrorCodeSignupBadRequest = ErrorCode("ERR_2101")
	ErrorCodeSignupConflict   = ErrorCode("ERR_2102")
)

type RenderableError struct {
	Code    ErrorCode
	Message string
	Details any
}

func (e *RenderableError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
