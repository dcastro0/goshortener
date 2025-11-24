package handlers

import (
	"encoding/base64"
	"fmt"
	"goshortener/internal/database"
	"goshortener/internal/models"
	"goshortener/pkg/utils"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ShortenRequest struct {
	URL       string `json:"url"`
	Alias     string `json:"alias"`
	Password  string `json:"password"`
	ExpiresAt string `json:"expires_at"`
}

type InspectRequest struct {
	Code     string `json:"code"`
	Password string `json:"password"`
}

type UpdateLinkRequest struct {
	OriginalURL string `json:"url"`
	Hash        string `json:"alias"`
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

	if req.Password != "" {
		bytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao processar senha"})
		}
		link.Password = string(bytes)
	}

	if req.ExpiresAt != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04", req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Formato de data inválido"})
		}
		link.ExpiresAt = &parsedTime
	} else {
		defaultExpire := time.Now().Add(30 * 24 * time.Hour)
		link.ExpiresAt = &defaultExpire
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

	response := map[string]interface{}{
		"original_url": link.OriginalURL,
		"hash":         link.Hash,
		"clicks":       link.Clicks,
		"created_at":   link.CreatedAt,
		"expires_at":   link.ExpiresAt,
		"protected":    false,
	}

	if link.Password != "" {
		response["protected"] = true
		if req.Password != "" {
			err := bcrypt.CompareHashAndPassword([]byte(link.Password), []byte(req.Password))
			if err == nil {
				response["unlocked"] = true
			} else {
				response["error"] = "Senha incorreta"
				response["original_url"] = "🔒 Protegido"
				response["clicks"] = -1
			}
		} else {
			response["original_url"] = "🔒 Protegido"
			response["clicks"] = -1
		}
	}

	return c.JSON(http.StatusOK, response)
}

func Redirect(c echo.Context) error {
	hash := c.Param("hash")

	var link models.ShortLink
	if err := database.DB.Where("hash = ?", hash).First(&link).Error; err != nil {
		return c.Render(http.StatusNotFound, "404", nil)
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return c.Render(http.StatusGone, "404", map[string]interface{}{
			"Title":   "Link Expirado",
			"Message": "Este link atingiu sua data de validade e não está mais disponível.",
		})
	}

	if link.Password != "" {
		if c.Request().Method == http.MethodPost {
			password := c.FormValue("password")
			err := bcrypt.CompareHashAndPassword([]byte(link.Password), []byte(password))
			if err != nil {
				return c.Render(http.StatusUnauthorized, "link_password", map[string]interface{}{
					"Hash":  hash,
					"Error": "Senha incorreta",
				})
			}
		} else {
			return c.Render(http.StatusOK, "link_password", map[string]interface{}{
				"Hash": hash,
			})
		}
	}

	database.DB.Model(&link).UpdateColumn("clicks", gorm.Expr("clicks + 1"))
	return c.Redirect(http.StatusFound, link.OriginalURL)
}

func UpdateLink(c echo.Context) error {
	id := c.Param("id")
	req := new(UpdateLinkRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}
	if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "URL inválida"})
	}
	var link models.ShortLink
	if err := database.DB.First(&link, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Link não encontrado"})
	}
	if req.Hash != link.Hash {
		var existing models.ShortLink
		if err := database.DB.Where("hash = ?", req.Hash).First(&existing).Error; err == nil {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Este alias já está em uso"})
		}
	}
	link.OriginalURL = req.OriginalURL
	link.Hash = req.Hash
	if err := database.DB.Save(&link).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao atualizar"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Link atualizado com sucesso!"})
}

func GetStats(c echo.Context) error {
	q := c.QueryParam("q")
	var links []models.ShortLink
	query := database.DB.Order("created_at desc").Limit(100)
	if q != "" {
		search := "%" + q + "%"
		query = query.Where("original_url ILIKE ? OR hash ILIKE ?", search, search)
	}
	if err := query.Find(&links).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar links"})
	}
	var messages []models.ContactMessage
	if err := database.DB.Order("created_at desc").Find(&messages).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao buscar mensagens"})
	}
	var totalLinks int64
	database.DB.Model(&models.ShortLink{}).Count(&totalLinks)
	var totalClicks int64
	database.DB.Model(&models.ShortLink{}).Select("coalesce(sum(clicks), 0)").Scan(&totalClicks)
	var totalMessages int64
	database.DB.Model(&models.ContactMessage{}).Count(&totalMessages)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memUsage := mem.Alloc / 1024 / 1024
	numGoroutines := runtime.NumGoroutine()
	return c.Render(http.StatusOK, "stats", map[string]interface{}{
		"FullWidth":     true,
		"Links":         links,
		"Messages":      messages,
		"Query":         q,
		"TotalLinks":    totalLinks,
		"TotalClicks":   totalClicks,
		"TotalMessages": totalMessages,
		"MemUsage":      memUsage,
		"NumGoroutines": numGoroutines,
		"GoVersion":     runtime.Version(),
	})
}

func DeleteLink(c echo.Context) error {
	id := c.Param("id")
	if err := database.DB.Delete(&models.ShortLink{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao deletar"})
	}
	return c.NoContent(http.StatusNoContent)
}

func DeleteMessage(c echo.Context) error {
	id := c.Param("id")
	if err := database.DB.Delete(&models.ContactMessage{}, id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao deletar mensagem"})
	}
	return c.NoContent(http.StatusNoContent)
}
