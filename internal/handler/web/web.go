package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

func NewRouter(conf config.Web, log *log.SugaredLogger) (*echo.Echo, error) {
	log.Info("Initializing web router")

	var (
		e *echo.Echo
	)
	e = echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	e.GET("/", HandleIndex)

	e.Static("/*.html", conf.StaticFilesPath)
	e.File("/favicon.ico", staticPath(conf, "assets/favicon.png"))
	if conf.APIHost != "" {
		e.GET("/assets/js/env.js", envJson{
			lastModified: time.Now().Format(http.TimeFormat),
			response:     fmt.Sprintf("const apiHost = '%s';", conf.APIHost),
		}.Handle)
	}
	e.Static("/assets", staticPath(conf, "assets"))

	e.HTTPErrorHandler = customHttpErrorHandler(conf, log)

	log.Info("Router initialized")
	return e, nil
}

type envJson struct {
	lastModified string
	response     string
}

func (e envJson) Handle(c echo.Context) error {
	if c.Request().Header.Get("If-Modified-Since") == e.lastModified {
		return c.NoContent(http.StatusNotModified)
	}
	c.Response().Header().Set("Last-Modified", e.lastModified)
	return c.String(http.StatusOK, e.response)
}

func HandleIndex(c echo.Context) error {
	err := c.Redirect(http.StatusFound, "/index.html")
	return err
}

func customHttpErrorHandler(conf config.Web, log *log.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			handleHttpError(httpError, c, log)
			return
		}

		log.Error("Unexpected error occurred: ", err)
		if err = c.File(staticPath(conf, "error.html")); err != nil {
			log.Error("Failed to serve error page: ", err)
		}
	}
}

func staticPath(conf config.Web, suffix string) string {
	return fmt.Sprintf("%s/%s", conf.StaticFilesPath, suffix)
}

func handleHttpError(httpError *echo.HTTPError, c echo.Context, log *log.SugaredLogger) {
	var redirectPage string
	switch httpError.Code {
	case http.StatusNotFound:
		redirectPage = "404"
	case http.StatusTooManyRequests:
		if sErr := c.NoContent(http.StatusTooManyRequests); sErr != nil {
			log.Error("Failed to serve too many requests error response: ", sErr)
			return
		}
	default:
		log.Error("Unexpected http error: ", httpError)
		redirectPage = "error"
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
