package handler

import (
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

func servePage(c echo.Context, page string, log *log.SugaredLogger) {
	if err := c.File(fmt.Sprintf("web/%s.html", page)); err != nil {
		log.Error("Failed to serve page %s: ", page, err)
	}
}
