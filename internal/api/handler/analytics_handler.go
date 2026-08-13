package handler

import (
	"net/http"
	"trendybackend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	useCase usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(useCase usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{useCase}
}

func (h *AnalyticsHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.useCase.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
