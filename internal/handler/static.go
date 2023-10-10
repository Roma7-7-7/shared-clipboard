package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "go.uber.org/zap"
)

func HandleIndex(c echo.Context) error {
	err := c.Redirect(http.StatusFound, "/index.html")
	return err
}

func HandleFavicon(c echo.Context) error {
	err := c.Redirect(http.StatusFound, "/assets/favicon.png")
	return err
}

func customHttpErrorHandler(log *log.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Error("Unexpected error occurred: ", err)

		if c.Path() != "/*.html*" {
			log.Error("Unexpected error occurred: ", err)
			if sErr := c.String(http.StatusInternalServerError, "Internal server error"); sErr != nil {
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

func servePage(c echo.Context, page string, log *log.SugaredLogger) {
	if err := c.File(fmt.Sprintf("web/%s.html", page)); err != nil {
		log.Error("Failed to serve page %s: ", page, err)
	}
}
