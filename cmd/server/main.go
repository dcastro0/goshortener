package main

import (
	"goshortener/internal/config"
	"goshortener/internal/database"
	"goshortener/internal/handlers"
	"goshortener/pkg/utils"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.Init()
	database.Init()

	e := echo.New()

	e.Static("/static", "static")

	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/index.html"))
	templates["stats"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/stats.html"))
	templates["link_password"] = template.Must(template.ParseFiles("internal/templates/link_password.html"))

	e.Renderer = &utils.TemplateRegistry{
		Templates: templates,
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})

	e.GET("/stats", handlers.GetStats)
	e.POST("/inspect", handlers.InspectLink) // Nova Rota
	e.POST("/shorten", handlers.ShortenURL)
	e.GET("/:hash", handlers.Redirect)

	e.Logger.Fatal(e.Start(":8080"))
}
