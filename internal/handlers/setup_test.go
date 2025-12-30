package handlers

import (
	"goshortener/internal/database"
	"goshortener/internal/models"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB() {
	var err error
	database.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Run migrations
	err = database.DB.AutoMigrate(&models.ShortLink{}, &models.ContactMessage{})
	if err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}
}

// SetupTestServer sets up an Echo instance for testing
func SetupTestServer() *echo.Echo {
	e := echo.New()

	// Register templates (mocking them since we don't render templates in API tests usually,
	// but good to have if we test handlers that render)
	templates := make(map[string]*template.Template)
	// We might need to mock TemplateRegistry if we test handlers that use c.Render
	e.Renderer = &TemplateRegistry{
		Templates: templates,
	}

	return e
}

// TemplateRegistry mock
type TemplateRegistry struct {
	Templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil // Mock implementation
}
