package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"trendybackend/internal/domain"
)

type ProductHandler struct {
	productUseCase domain.ProductUseCase
}

func NewProductHandler(u domain.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: u,
	}
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	search := c.Query("search")
	categoryIDStr := c.Query("category_id")
	subCategoryIDStr := c.Query("sub_category_id")

	var categoryID uint
	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = uint(id)
		}
	}

	var subCategoryID uint
	if subCategoryIDStr != "" {
		if id, err := strconv.Atoi(subCategoryIDStr); err == nil {
			subCategoryID = uint(id)
		}
	}

	products, err := h.productUseCase.GetAllProducts(search, categoryID, subCategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	product, err := h.productUseCase.GetProduct(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var product domain.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.productUseCase.CreateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var product domain.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	product.ID = uint(id)

	if err := h.productUseCase.UpdateProduct(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.productUseCase.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}
func (h *ProductHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Create unique filename
	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	log.Printf("Received upload request for file: %s, saving as: %s", file.Filename, newFileName)

	// Resolve uploads directory (consistent with main.go)
	uploadDir := os.Getenv("UPLOADS_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Printf("Error creating upload directory %s: %v", uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create uploads directory"})
		return
	}

	// Save file
	path := filepath.Join(uploadDir, newFileName)
	log.Printf("Saving file to path: %s", path)
	if err := c.SaveUploadedFile(file, path); err != nil {
		log.Printf("Error saving file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
		return
	}

	// Return the relative URL
	url := fmt.Sprintf("/uploads/%s", newFileName)
	log.Printf("Upload successful. Returning URL: %s", url)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// Helper to handle general file uploads
func (h *ProductHandler) uploadFileHelper(c *gin.Context, formName string) (string, error) {
	file, err := c.FormFile(formName)
	if err != nil {
		return "", err
	}
	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	uploadDir := os.Getenv("UPLOADS_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}
	path := filepath.Join(uploadDir, newFileName)
	if err := c.SaveUploadedFile(file, path); err != nil {
		return "", err
	}
	return fmt.Sprintf("/uploads/%s", newFileName), nil
}

// Create Variant
func (h *ProductHandler) CreateVariant(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var variant domain.ProductVariant
	if err := c.ShouldBindJSON(&variant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	variant.ProductID = uint(productID)

	if err := h.productUseCase.CreateVariant(&variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, variant)
}

// Update Variant
func (h *ProductHandler) UpdateVariant(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var variant domain.ProductVariant
	if err := c.ShouldBindJSON(&variant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	variant.ID = uint(variantID)

	// Keep ProductID from URL if not specified
	if variant.ProductID == 0 {
		productID, _ := strconv.Atoi(c.Param("id"))
		variant.ProductID = uint(productID)
	}

	if err := h.productUseCase.UpdateVariant(&variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, variant)
}

// Delete Variant
func (h *ProductHandler) DeleteVariant(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	if err := h.productUseCase.DeleteVariant(uint(variantID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Variant deleted"})
}

// Upload Variant Images
func (h *ProductHandler) UploadVariantImages(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	url, err := h.uploadFileHelper(c, "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isPrimaryStr := c.PostForm("is_primary")
	isPrimary, _ := strconv.ParseBool(isPrimaryStr)
	sortOrderStr := c.PostForm("sort_order")
	sortOrder, _ := strconv.Atoi(sortOrderStr)

	img := domain.VariantImage{
		VariantID: uint(variantID),
		Image:     url,
		ImageURL:  url,
		IsPrimary: isPrimary,
		SortOrder: sortOrder,
	}

	if err := h.productUseCase.AddVariantImage(&img); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

// Upload Variant Video
func (h *ProductHandler) UploadVariantVideo(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	url, err := h.uploadFileHelper(c, "video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.VideoURL = url
	if err := h.productUseCase.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"video_url": url})
}

// Upload Variant PDF
func (h *ProductHandler) UploadVariantPDF(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	url, err := h.uploadFileHelper(c, "pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.PdfURL = url
	if err := h.productUseCase.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pdf_url": url})
}

// PATCH Update Stock
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	type StockReq struct {
		Stock int `json:"stock"`
	}

	var req StockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.Stock = req.Stock
	if err := h.productUseCase.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// PATCH Update Price
func (h *ProductHandler) UpdatePrice(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	type PriceReq struct {
		Price     float64 `json:"price"`
		SalePrice float64 `json:"sale_price"`
	}

	var req PriceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.Price = req.Price
	if req.SalePrice > 0 {
		variant.SalePrice = req.SalePrice
	}
	if err := h.productUseCase.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// PATCH Update Expiry Date
func (h *ProductHandler) UpdateExpiry(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	type ExpiryReq struct {
		ExpiryDate string `json:"expiry_date"`
	}

	var req ExpiryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.ExpiryDate = req.ExpiryDate
	if err := h.productUseCase.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// Create Variant Review
func (h *ProductHandler) CreateReview(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var review domain.VariantReview
	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	review.VariantID = uint(variantID)
	review.CreatedAt = time.Now()

	if err := h.productUseCase.AddVariantReview(&review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// Duplicate Variant
func (h *ProductHandler) DuplicateVariant(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	v, err := h.productUseCase.GetVariantByID(uint(variantID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	newVariant := domain.ProductVariant{
		ProductID:        v.ProductID,
		Title:            v.Title + " (Copy)",
		TitleAr:          v.TitleAr,
		ShortDescription: v.ShortDescription,
		Description:      v.Description,
		DescriptionAr:    v.DescriptionAr,
		Weight:           v.Weight,
		PackageType:      v.PackageType,
		Texture:          v.Texture,
		ExpiryDate:       v.ExpiryDate,
		SKU:              v.SKU + "-COPY",
		Barcode:          v.Barcode,
		Price:            v.Price,
		SalePrice:        v.SalePrice,
		CostPrice:        v.CostPrice,
		Stock:            v.Stock,
		LowStockAlert:    v.LowStockAlert,
		Status:           v.Status,
		VideoURL:         v.VideoURL,
		PdfURL:           v.PdfURL,
	}

	for _, img := range v.Images {
		newVariant.Images = append(newVariant.Images, domain.VariantImage{
			Image:     img.Image,
			ImageURL:  img.ImageURL,
			SortOrder: img.SortOrder,
			IsPrimary: img.IsPrimary,
		})
	}

	if err := h.productUseCase.CreateVariant(&newVariant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newVariant)
}

// Bulk Import Variants
func (h *ProductHandler) BulkImport(c *gin.Context) {
	type ImportRequest struct {
		ProductID        uint    `json:"product_id"`
		Title            string  `json:"title"`
		TitleAr          string  `json:"title_ar"`
		ShortDescription string  `json:"short_description"`
		Description      string  `json:"description"`
		DescriptionAr    string  `json:"description_ar"`
		Weight           float64 `json:"weight"`
		PackageType      string  `json:"package_type"`
		Texture          string  `json:"texture"`
		ExpiryDate       string  `json:"expiry_date"`
		SKU              string  `json:"sku"`
		Barcode          string  `json:"barcode"`
		Price            float64 `json:"price"`
		SalePrice        float64 `json:"sale_price"`
		CostPrice        float64 `json:"cost_price"`
		Stock            int     `json:"stock"`
		LowStockAlert    int     `json:"low_stock_alert"`
		Status           string  `json:"status"`
	}

	var reqs []ImportRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	importedCount := 0
	for _, req := range reqs {
		if _, err := h.productUseCase.GetProduct(req.ProductID); err != nil {
			continue
		}

		v := domain.ProductVariant{
			ProductID:        req.ProductID,
			Title:            req.Title,
			TitleAr:          req.TitleAr,
			ShortDescription: req.ShortDescription,
			Description:      req.Description,
			DescriptionAr:    req.DescriptionAr,
			Weight:           req.Weight,
			PackageType:      req.PackageType,
			Texture:          req.Texture,
			ExpiryDate:       req.ExpiryDate,
			SKU:              req.SKU,
			Barcode:          req.Barcode,
			Price:            req.Price,
			SalePrice:        req.SalePrice,
			CostPrice:        req.CostPrice,
			Stock:            req.Stock,
			LowStockAlert:    req.LowStockAlert,
			Status:           req.Status,
		}
		if v.Status == "" {
			v.Status = "Active"
		}
		if err := h.productUseCase.CreateVariant(&v); err == nil {
			importedCount++
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully imported %d variants", importedCount)})
}

// Bulk Export Variants
func (h *ProductHandler) BulkExport(c *gin.Context) {
	products, err := h.productUseCase.GetAllProducts("", 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type ExportedVariant struct {
		ParentProductID uint    `json:"parent_product_id"`
		ParentName      string  `json:"parent_name"`
		VariantID       uint    `json:"variant_id"`
		Title           string  `json:"title"`
		TitleAr         string  `json:"title_ar"`
		SKU             string  `json:"sku"`
		Barcode         string  `json:"barcode"`
		Price           float64 `json:"price"`
		SalePrice       float64 `json:"sale_price"`
		CostPrice       float64 `json:"cost_price"`
		Stock           int     `json:"stock"`
		Weight          float64 `json:"weight"`
		PackageType     string  `json:"package_type"`
		Texture         string  `json:"texture"`
		ExpiryDate      string  `json:"expiry_date"`
		Status          string  `json:"status"`
	}

	var exported []ExportedVariant
	for _, p := range products {
		for _, v := range p.Variants {
			exported = append(exported, ExportedVariant{
				ParentProductID: p.ID,
				ParentName:      p.Name,
				VariantID:       v.ID,
				Title:           v.Title,
				TitleAr:         v.TitleAr,
				SKU:             v.SKU,
				Barcode:         v.Barcode,
				Price:           v.Price,
				SalePrice:       v.SalePrice,
				CostPrice:       v.CostPrice,
				Stock:           v.Stock,
				Weight:          v.Weight,
				PackageType:     v.PackageType,
				Texture:         v.Texture,
				ExpiryDate:      v.ExpiryDate,
				Status:          v.Status,
			})
		}
	}
	c.JSON(http.StatusOK, exported)
}
