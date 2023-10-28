package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type AppContext struct {
	echo.Context
}

func (c *AppContext) Request() *http.Request {
	tid := "undefined"
	if c.Response().Header().Get(echo.HeaderXRequestID) != "" {
		tid = c.Response().Header().Get(echo.HeaderXRequestID)
	}
	return c.Request().WithContext(trace.WithTraceID(c.Request().Context(), tid))
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(&AppContext{c})
	}
}
