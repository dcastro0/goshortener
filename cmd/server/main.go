package main

import (
	"crypto/subtle"
	"goshortener/internal/config"
	"goshortener/internal/database"
	"goshortener/internal/handlers"
	"goshortener/internal/models"
	"goshortener/pkg/utils"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func main() {
	config.Init()
	database.Init()

	database.DB.AutoMigrate(&models.ContactMessage{})

	e := echo.New()

	e.Static("/static", "static")

	templates := make(map[string]*template.Template)
	templates["index"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/index.html"))
	templates["stats"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/stats.html"))
	templates["link_password"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/link_password.html"))
	templates["404"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/404.html"))
	templates["terms"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/terms.html"))
	templates["contact"] = template.Must(template.ParseFiles("internal/templates/base.html", "internal/templates/contact.html"))

	e.Renderer = &utils.TemplateRegistry{
		Templates: templates,
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	e.GET("/robots.txt", func(c echo.Context) error {
		return c.String(http.StatusOK, "User-agent: *\nAllow: /\nDisallow: /stats\nDisallow: /link/")
	})

	e.GET("/sitemap.xml", func(c echo.Context) error {
		sitemap := `<?xml version="1.0" encoding="UTF-8"?>
		<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			<url><loc>https://goshortener.com/</loc><changefreq>daily</changefreq></url>
			<url><loc>https://goshortener.com/terms</loc><changefreq>monthly</changefreq></url>
			<url><loc>https://goshortener.com/contact</loc><changefreq>monthly</changefreq></url>
		</urlset>`
		return c.XMLBlob(http.StatusOK, []byte(sitemap))
	})

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", nil)
	})

	e.GET("/terms", func(c echo.Context) error { return c.Render(http.StatusOK, "terms", nil) })
	e.GET("/contact", func(c echo.Context) error { return c.Render(http.StatusOK, "contact", nil) })
	e.POST("/contact", handlers.SendContact)

	admin := e.Group("", middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		envUser := viper.GetString("ADMIN_USER")
		envPass := viper.GetString("ADMIN_PASSWORD")
		if subtle.ConstantTimeCompare([]byte(username), []byte(envUser)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(envPass)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	admin.GET("/stats", handlers.GetStats)
	admin.DELETE("/link/:id", handlers.DeleteLink)
	admin.DELETE("/message/:id", handlers.DeleteMessage)

	e.POST("/inspect", handlers.InspectLink)
	e.POST("/shorten", handlers.ShortenURL)
	e.Match([]string{http.MethodGet, http.MethodPost}, "/:hash", handlers.Redirect)

	e.Logger.Fatal(e.Start(":8080"))
}
