package handler

import (
	"github.com/labstack/echo/v4"
	log "go.uber.org/zap"
)

func customHttpErrorHandler(log *log.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Error("Unexpected error occurred: ", err)

		if fErr := c.File("web/static/error.html"); fErr != nil {
			log.Error("Failed to serve error page: ", fErr)
		}
	}
}
