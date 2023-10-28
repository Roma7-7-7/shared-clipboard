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
	original := c.Context.Request()
	tid := "undefined"
	if c.Response().Header().Get(echo.HeaderXRequestID) != "" {
		tid = c.Response().Header().Get(echo.HeaderXRequestID)
	}
	return original.WithContext(trace.WithTraceID(original.Context(), tid))
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(&AppContext{c})
	}
}
