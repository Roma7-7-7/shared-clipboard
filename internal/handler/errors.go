package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "go.uber.org/zap"
)

type errorCode string

var (
	errCodeInternalServer  = errorCode("ERROR_000")
	errorCodeCreateSession = errorCode("ERROR_100")
)

type APIError struct {
	RootError error     `json:"-"`
	Code      errorCode `json:"code"`
	Message   string    `json:"message"`
	Details   any       `json:"details"`
}

func NewAPIError(err error, code errorCode, message string, details any) APIError {
	return APIError{
		RootError: err,
		Code:      code,
		Message:   message,
		Details:   details,
	}
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Code, e.Message, e.RootError)
}

func customHttpErrorHandler(log *log.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Error("Unexpected error occurred: ", err)

		// TODO doesn't work with api/random-path
		if c.Path() != "/*.html*" {
			log.Error("Unexpected error occurred: ", err)
			if sErr := c.String(http.StatusInternalServerError, "Internal server error"); sErr != nil {
				log.Error("Failed to serve error response: ", sErr)
			}

			return
		}

		var apiError *APIError
		if errors.As(err, &apiError) {
			log.Error("API error occurred: ", apiError)
			if sErr := c.JSON(http.StatusInternalServerError, apiError); sErr != nil {
				log.Error("Failed to serve error response: ", sErr)
			}

			return
		}

		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			if httpError.Code == http.StatusNotFound {
				servePage(c, "404", log)
				return
			}
		}

		// do not redirect to not cause infinite loop in case error constantly occurs
		servePage(c, "error", log)
	}
}
