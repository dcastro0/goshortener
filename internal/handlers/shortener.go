package handlers

import (
	"encoding/base64"
	"fmt"
	"goshortener/internal/database"
	"goshortener/internal/models"
	"goshortener/pkg/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type InspectRequest struct {
	Code string `json:"code"`
}

func ShortenURL(c echo.Context) error {
	req := new(ShortenRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Requisição inválida"})
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "URL inválida"})
	}

	var hash string

	if req.Alias != "" {
		var existing models.ShortLink
		if err := database.DB.Where("hash = ?", req.Alias).First(&existing).Error; err == nil {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Este alias já está em uso"})
		}
		hash = req.Alias
	} else {
		hash = utils.GenerateRandomString(6)
	}

	link := models.ShortLink{
		OriginalURL: req.URL,
		Hash:        hash,
	}

	if err := database.DB.Create(&link).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar no banco"})
	}

	fullShortURL := fmt.Sprintf("http://%s/%s", c.Request().Host, hash)

	png, _ := qrcode.Encode(fullShortURL, qrcode.Medium, 256)

	qrBase64 := ""
	if png != nil {
		qrBase64 = base64.StdEncoding.EncodeToString(png)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "Link encurtado com sucesso",
		"short_url": c.Request().Host + "/" + link.Hash,
		"hash":      hash,
		"qr_code":   qrBase64,
	})
}

func InspectLink(c echo.Context) error {
	req := new(InspectRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Requisição inválida"})
	}

	hash := req.Code
	if strings.Contains(hash, "/") {
		parts := strings.Split(hash, "/")
		hash = parts[len(parts)-1]
	}

	var link models.ShortLink
	if err := database.DB.Where("hash = ?", hash).First(&link).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Link não encontrado"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"original_url": link.OriginalURL,
		"hash":         link.Hash,
		"clicks":       link.Clicks,
		"created_at":   link.CreatedAt,
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
