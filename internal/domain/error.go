package domain

import (
	"fmt"
	"net/http"
)

type (
	ErrorCode struct {
		Value      string
		StatusCode int
	}
)

var (
	ErrNotFound = fmt.Errorf("not found")

	ErrorBadRequest              = ErrorCode{"ERR_0400", http.StatusBadRequest}
	ErrorCodeUnauthorized        = ErrorCode{"ERR_0401", http.StatusUnauthorized}
	ErrorCodeForbidden           = ErrorCode{"ERR_0403", http.StatusForbidden}
	ErrorCodeNotFound            = ErrorCode{"ERR_0404", http.StatusNotFound}
	ErrorCodeMethodNotAllowed    = ErrorCode{"ERR_0405", http.StatusMethodNotAllowed}
	ErrorCodeInternalServerError = ErrorCode{"ERR_0500", http.StatusInternalServerError}

	ErrorCodeSignupBadRequest   = ErrorCode{"ERR_2101", http.StatusBadRequest}
	ErrorCodeSignupConflict     = ErrorCode{"ERR_2102", http.StatusBadRequest}
	ErrorCodeSiginWrongPassword = ErrorCode{"ERR_2103", http.StatusForbidden}

	ErrorCodeUserNotFound = ErrorCode{"ERR_2201", http.StatusBadRequest}
)

type RenderableError struct {
	Code    ErrorCode
	Message string
	Details any
}

func (e *RenderableError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code.Value, e.Message)
}
