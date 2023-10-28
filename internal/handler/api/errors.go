package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type errorCode string

var (
	// Generic error codes
	errorCodeServerError = errorCode("ERR_0500")

	// Session error codes
	errorCodeCreateSession = errorCode("ERR_1000")
)

func customHttpErrorHandler(log trace.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Errorw(c.Request().Context(), "Unexpected error occurred", err)

		var apiError *Error
		if errors.As(err, &apiError) {
			if sErr := c.JSON(apiError.HTTPStatus, apiError.Response); sErr != nil {
				log.Errorw(c.Request().Context(), "Failed to serve error response", sErr)
			}
			return
		}

		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			if sErr := c.JSON(httpError.Code, Response{
				Error:   true,
				Code:    fmt.Sprintf("ERR_0%-3d", httpError.Code),
				Message: http.StatusText(httpError.Code),
			}); sErr != nil {
				log.Errorw(c.Request().Context(), "Failed to serve error response", sErr)
			}
			return
		}

		if sErr := c.JSON(http.StatusInternalServerError, Response{
			Error:   true,
			Code:    string(errorCodeServerError),
			Message: http.StatusText(http.StatusInternalServerError),
		}); sErr != nil {
			log.Errorw(c.Request().Context(), "failed to server error response", sErr)
		}
	}
}
