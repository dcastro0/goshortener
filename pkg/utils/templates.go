package utils

import (
	"errors"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type TemplateRegistry struct {
	Templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.Templates[name]
	if !ok {
		return errors.New("Template not found -> " + name)
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}
