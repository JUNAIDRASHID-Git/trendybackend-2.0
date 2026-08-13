package handler

import (
	"net/http"
	"trendybackend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	useCase usecase.NotificationUseCase
}

func NewNotificationHandler(u usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{useCase: u}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	notifications, err := h.useCase.GetNotifications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}
