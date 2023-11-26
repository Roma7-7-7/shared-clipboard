package domain

import "fmt"

type (
	ErrorCode string
)

var (
	ErrorBadRequest              = ErrorCode("ERR_0400")
	ErrorCodeUnauthorized        = ErrorCode("ERR_0401")
	ErrorCodeForbidden           = ErrorCode("ERR_0403")
	ErrorCodeNotFound            = ErrorCode("ERR_0404")
	ErrorCodeMethodNotAllowed    = ErrorCode("ERR_0405")
	ErrorCodeInternalServerError = ErrorCode("ERR_0500")

	ErrorCodeMarshalResponse = ErrorCode("ERR_1001")

	ErrorCodeSignupBadRequest   = ErrorCode("ERR_2101")
	ErrorCodeSignupConflict     = ErrorCode("ERR_2102")
	ErrorCodeSiginWrongPassword = ErrorCode("ERR_2103")

	ErrorCodeUserNotFound = ErrorCode("ERR_2201")
)

type RenderableError struct {
	Code    ErrorCode
	Message string
	Details any
}

func (e *RenderableError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
