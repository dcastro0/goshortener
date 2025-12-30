package handlers

import (
	"encoding/json"
	"goshortener/internal/database"
	"goshortener/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestShortenURL(t *testing.T) {
	SetupTestDB()
	e := SetupTestServer()

	t.Run("Success - Random Alias", func(t *testing.T) {
		reqBody := `{"url": "https://example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, ShortenURL(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var res map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)

			assert.Equal(t, "Link encurtado com sucesso", res["message"])
			assert.NotEmpty(t, res["hash"])
			assert.NotEmpty(t, res["short_url"])

			// Verify DB
			var link models.ShortLink
			err = database.DB.Where("hash = ?", res["hash"]).First(&link).Error
			assert.NoError(t, err)
			assert.Equal(t, "https://example.com", link.OriginalURL)
		}
	})

	t.Run("Success - Custom Alias", func(t *testing.T) {
		reqBody := `{"url": "https://google.com", "alias": "google"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, ShortenURL(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var res map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &res)
			assert.Equal(t, "google", res["hash"])
		}
	})

	t.Run("Error - Alias Already Exists", func(t *testing.T) {
		// First create the link
		database.DB.Create(&models.ShortLink{OriginalURL: "https://old.com", Hash: "existing"})

		reqBody := `{"url": "https://new.com", "alias": "existing"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, ShortenURL(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})

	t.Run("Error - Invalid URL", func(t *testing.T) {
		reqBody := `{"url": "not-a-url"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, ShortenURL(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}
