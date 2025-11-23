package handlers

import (
	"goshortener/internal/database"
	"goshortener/internal/models"
	"goshortener/pkg/utils"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

func ShortenURL(c echo.Context) error {
	req := new(ShortenRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Requisição inválida"})
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "URL inválida"})
	}

	hash := utils.GenerateRandomString(6)
	link := models.ShortLink{
		OriginalURL: req.URL,
		Hash:        hash,
	}

	if err := database.DB.Create(&link).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar no banco"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "Link encurtado com sucesso",
		"short_url": c.Request().Host + "/" + link.Hash,
		"hash":      hash,
	})
}

func Redirect(c echo.Context) error {
	hash := c.Param("hash")

	var link models.ShortLink

	if err := database.DB.Where("hash = ?", hash).First(&link).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Link não encontrado"})
	}

	database.DB.Model(&link).UpdateColumn("clicks", gorm.Expr("clicks + 1"))

	return c.Redirect(http.StatusFound, link.OriginalURL)
}

func GetStats(c echo.Context) error {
	var links []models.ShortLink
	if err := database.DB.Order("created_at desc").Find(&links).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar dados"})
	}

	return c.Render(http.StatusOK, "stats", map[string]interface{}{
		"Links": links,
	})
}
