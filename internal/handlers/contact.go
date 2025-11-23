package handlers

import (
	"goshortener/internal/database"
	"goshortener/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func SendContact(c echo.Context) error {
	req := new(ContactRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	msg := models.ContactMessage{
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Message: req.Message,
	}

	if err := database.DB.Create(&msg).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Erro ao salvar mensagem"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Mensagem enviada com sucesso!",
	})
}
