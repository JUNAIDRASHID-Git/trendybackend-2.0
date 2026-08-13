package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"trendybackend/internal/domain"
)

// ── Category Handler ─────────────────────────────────────────────────────────

type CategoryHandler struct {
	categoryUseCase    domain.CategoryUseCase
	subCategoryUseCase domain.SubCategoryUseCase
}

func NewCategoryHandler(cu domain.CategoryUseCase, su domain.SubCategoryUseCase) *CategoryHandler {
	return &CategoryHandler{categoryUseCase: cu, subCategoryUseCase: su}
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.categoryUseCase.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	category, err := h.categoryUseCase.GetCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category domain.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.categoryUseCase.CreateCategory(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var category domain.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category.ID = uint(id)
	if err := h.categoryUseCase.UpdateCategory(&category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.categoryUseCase.DeleteCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// ── SubCategory Handler ──────────────────────────────────────────────────────

func (h *CategoryHandler) GetAllSubCategories(c *gin.Context) {
	// Optional filter by category_id query param
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		catID, err := strconv.Atoi(catIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_id"})
			return
		}
		subs, err := h.subCategoryUseCase.GetSubCategoriesByCategoryID(uint(catID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, subs)
		return
	}
	subs, err := h.subCategoryUseCase.GetAllSubCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *CategoryHandler) GetSubCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	sub, err := h.subCategoryUseCase.GetSubCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SubCategory not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *CategoryHandler) CreateSubCategory(c *gin.Context) {
	var sub domain.SubCategory
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.subCategoryUseCase.CreateSubCategory(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (h *CategoryHandler) UpdateSubCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var sub domain.SubCategory
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub.ID = uint(id)
	if err := h.subCategoryUseCase.UpdateSubCategory(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *CategoryHandler) DeleteSubCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.subCategoryUseCase.DeleteSubCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "SubCategory deleted"})
}
