package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"trendybackend/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UIHandler struct {
	db *gorm.DB
}

func NewUIHandler(db *gorm.DB) *UIHandler {
	return &UIHandler{db: db}
}

// UploadHeroVideo accepts a multipart MP4 upload, saves it to uploads/ui/,
// and persists its public URL as the "hero_video_url" setting.
func (h *UIHandler) UploadHeroVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided: " + err.Error()})
		return
	}

	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}

	uiDir := filepath.Join(uploadsDir, "ui")
	if err := os.MkdirAll(uiDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ui uploads directory"})
		return
	}

	// Always use a fixed filename so re-uploading replaces the old video
	destPath := filepath.Join(uiDir, "hero_video.mp4")

	// Remove old file if exists
	os.Remove(destPath)

	if err := c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video: " + err.Error()})
		return
	}

	// Build the public URL
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	videoURL := fmt.Sprintf("%s://%s/uploads/ui/hero_video.mp4", scheme, host)

	// Persist as a setting
	h.saveSetting("hero_video_url", videoURL)

	c.JSON(http.StatusOK, gin.H{
		"url":     videoURL,
		"message": "Hero video uploaded successfully",
	})
}

// DeleteHeroVideo removes the hero video file and clears the hero_video_url setting.
func (h *UIHandler) DeleteHeroVideo(c *gin.Context) {
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}

	destPath := filepath.Join(uploadsDir, "ui", "hero_video.mp4")
	os.Remove(destPath) // Ignore error if file doesn't exist

	// Clear the setting
	h.saveSetting("hero_video_url", "")

	c.JSON(http.StatusOK, gin.H{"message": "Hero video deleted successfully"})
}

// saveSetting is a helper that upserts a key/value setting in the DB.
func (h *UIHandler) saveSetting(key, value string) {
	var setting domain.Setting
	err := h.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = domain.Setting{Key: key, Value: value}
			h.db.Create(&setting)
		}
	} else {
		setting.Value = value
		h.db.Save(&setting)
	}
}
