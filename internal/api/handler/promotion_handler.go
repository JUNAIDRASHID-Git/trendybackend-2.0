package handler

import (
	"net/http"
	"strconv"
	"trendybackend/internal/domain"

	"github.com/gin-gonic/gin"
)

type PromotionHandler struct {
	useCase domain.PromotionUseCase
}

func NewPromotionHandler(useCase domain.PromotionUseCase) *PromotionHandler {
	return &PromotionHandler{useCase: useCase}
}

func (h *PromotionHandler) GetAll(c *gin.Context) {
	promotions, err := h.useCase.GetAllPromotions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, promotions)
}

func (h *PromotionHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	promotion, err := h.useCase.GetPromotion(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Promotion not found"})
		return
	}
	c.JSON(http.StatusOK, promotion)
}

func (h *PromotionHandler) Create(c *gin.Context) {
	var promotion domain.Promotion
	if err := c.ShouldBindJSON(&promotion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreatePromotion(&promotion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, promotion)
}

func (h *PromotionHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var promotion domain.Promotion
	if err := c.ShouldBindJSON(&promotion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	promotion.ID = uint(id)

	if err := h.useCase.UpdatePromotion(&promotion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, promotion)
}

func (h *PromotionHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.useCase.DeletePromotion(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Promotion deleted"})
}
