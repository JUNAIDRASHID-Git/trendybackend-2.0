package handler

import (
	"net/http"
	"trendybackend/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingHandler struct {
	db *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{db: db}
}

func (h *SettingHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	var setting domain.Setting
	if err := h.db.Where("key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *SettingHandler) SaveSetting(c *gin.Context) {
	var input struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var setting domain.Setting
	err := h.db.Where("key = ?", input.Key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = domain.Setting{
				Key:   input.Key,
				Value: input.Value,
			}
			if err := h.db.Create(&setting).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		setting.Value = input.Value
		if err := h.db.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, setting)
}
