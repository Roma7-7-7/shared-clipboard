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
	// Generic error codes
	errorCodeServerError = errorCode("ERR_0500")

	// Session error codes
	errorCodeCreateSession = errorCode("ERR_1000")
)

func customHttpErrorHandler(log *log.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Error("Unexpected error occurred: ", err)

		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			handleHttpError(httpError, c, log)
			return
		}

		if err = c.File("web/error.html"); err != nil {
			log.Error("Failed to serve error page: ", err)
		}
	}
}

func handleAPIError(c *APIContext, err error, log *log.SugaredLogger) {
	var apiError *APIError
	if errors.As(err, &apiError) {
		if sErr := c.JSON(apiError.HTTPStatus, apiError.Response); sErr != nil {
			log.Error("Failed to serve APIError response: ", sErr)
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
			log.Error("Failed to serve API HTTPError response: ", sErr)
		}
		return
	}

	if sErr := c.JSON(http.StatusInternalServerError, Response{
		Error:   true,
		Code:    string(errorCodeServerError),
		Message: http.StatusText(http.StatusInternalServerError),
	}); sErr != nil {
		log.Error("Failed to serve API error response: ", sErr)
	}
}

func handleHttpError(httpError *echo.HTTPError, c echo.Context, log *log.SugaredLogger) {
	redirectPage := "error"
	switch httpError.Code {
	case http.StatusNotFound:
		redirectPage = "404"
	case http.StatusTooManyRequests:
		if sErr := c.NoContent(http.StatusTooManyRequests); sErr != nil {
			log.Error("Failed to serve too many requests error response: ", sErr)
			return
		}
	}

	redirectPage = fmt.Sprintf("/%s.html", redirectPage)

	if redirectPage == c.Request().URL.Path { // to prevent infinite redirect loop
		if sErr := c.String(http.StatusInternalServerError, "Internal server error"); sErr != nil {
			log.Error("Failed to serve error response in infinite redirects catch loop: ", sErr)
			return
		}
		return
	}

	if err := c.Redirect(http.StatusFound, redirectPage); err != nil {
		log.Error("Failed to redirect to page %s: ", redirectPage, err)
	}
}
