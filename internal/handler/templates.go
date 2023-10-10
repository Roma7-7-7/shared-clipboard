package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type TemplatesRenderer struct {
	templates *template.Template
}

func NewTemplatesRenderer(glob string) (*TemplatesRenderer, error) {
	tmpl, err := template.ParseGlob(glob)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &TemplatesRenderer{
		templates: tmpl,
	}, nil
}

func (r *TemplatesRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	var buff bytes.Buffer
	if err := r.templates.ExecuteTemplate(&buff, name, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	if _, err := io.Copy(w, &buff); err != nil {
		return fmt.Errorf("copy rendered template to writer: %w", err)
	}
	return nil
}

func HandleIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"msg": "Hello, World!",
	})
}
