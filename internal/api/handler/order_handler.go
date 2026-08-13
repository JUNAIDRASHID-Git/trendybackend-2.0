package handler

import (
	"net/http"
	"strconv"
	"trendybackend/internal/api/websocket"
	"trendybackend/internal/domain"
	"trendybackend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	useCase usecase.OrderUseCase
}

func NewOrderHandler(useCase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase}
}

func (h *OrderHandler) GetAll(c *gin.Context) {
	// Allow filtering by customer email for user-side order tracking
	email := c.Query("email")
	if email != "" {
		orders, err := h.useCase.GetByEmail(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
		return
	}

	orders, err := h.useCase.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	order, err := h.useCase.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) Create(c *gin.Context) {
	var order domain.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.Status == "" {
		order.Status = "Pending"
	}
	if order.OrderType == "" {
		order.OrderType = "Delivery"
	}

	if err := h.useCase.Create(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast the new order event via WebSocket!
	websocket.GetHub().Broadcast("new_order", order)

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var order domain.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.ID = uint(id)
	if err := h.useCase.Update(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update via WebSocket so frontend can track live
	websocket.GetHub().Broadcast("order_updated", order)

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.useCase.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}
