package domain

import (
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	Name          string           `gorm:"not null" json:"name"` // Product Group Name
	NameAr        string           `json:"name_ar"`
	CategoryID    uint             `json:"category_id"`
	SubCategoryID uint             `json:"sub_category_id"`
	BrandID       uint             `json:"brand_id"`
	Brand         string           `json:"brand"`
	Status        string           `gorm:"default:'Active'" json:"status"`
	// Fallback fields for compatibility
	Price         float64          `gorm:"not null" json:"price"`
	ImageURL      string           `json:"image_url"`
	Stock         int              `gorm:"default:0" json:"stock"`
	Sales         int              `gorm:"default:0" json:"sales"`
	IsTrending    bool             `gorm:"default:false" json:"is_trending"`
	Tags          []Tag            `gorm:"many2many:product_tags;" json:"tags"`
	Variants      []ProductVariant `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"variants"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

type ProductVariant struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	ProductID        uint            `gorm:"index" json:"product_id"`
	Title            string          `gorm:"default:'';not null" json:"title"` // Variant Specific Title
	TitleAr          string          `json:"title_ar"`
	ShortDescription string          `json:"short_description"`
	Description      string          `json:"description"`
	DescriptionAr    string          `json:"description_ar"`
	Weight           float64         `gorm:"default:0" json:"weight"`
	PackageType      string          `json:"package_type"`
	Texture          string          `json:"texture"`
	ExpiryDate       string          `json:"expiry_date"`
	SKU              string          `json:"sku"`
	Barcode          string          `json:"barcode"`
	Price            float64         `gorm:"default:0;not null" json:"price"`
	SalePrice        float64         `gorm:"default:0" json:"sale_price"`
	CostPrice        float64         `gorm:"default:0" json:"cost_price"`
	Stock            int             `gorm:"default:0" json:"stock"`
	LowStockAlert    int             `gorm:"default:5" json:"low_stock_alert"`
	AverageRating    float64         `gorm:"default:0" json:"average_rating"`
	ReviewCount      int             `gorm:"default:0" json:"review_count"`
	Slug             string          `json:"slug"`
	MetaTitle        string          `json:"meta_title"`
	MetaDescription  string          `json:"meta_description"`
	Keywords         string          `json:"keywords"`
	Status           string          `gorm:"default:'Active'" json:"status"` // Active, Draft, Out of Stock
	IsDefault        bool            `gorm:"default:false" json:"is_default"`
	VideoURL         string          `json:"video_url"`
	PdfURL           string          `json:"pdf_url"`
	Images           []VariantImage  `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"images"`
	Reviews          []VariantReview `gorm:"foreignKey:VariantID;constraint:OnDelete:CASCADE" json:"reviews"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

type VariantImage struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	VariantID uint   `gorm:"index" json:"variant_id"`
	Image     string `json:"image"`     // DB Column: image
	ImageURL  string `json:"image_url"` // Alias for frontend
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsPrimary bool   `gorm:"default:false" json:"is_primary"`
}

type VariantReview struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	VariantID    uint      `gorm:"index" json:"variant_id"`
	CustomerID   uint      `gorm:"index" json:"customer_id"`
	CustomerName string    `gorm:"type:varchar(255)" json:"customer_name"`
	Rating       int       `json:"rating"`
	Review       string    `json:"review"`
	CreatedAt    time.Time `json:"created_at"`
}


func (p *Product) BeforeSave(tx *gorm.DB) (err error) {
	if len(p.Variants) > 0 {
		var defaultVariant *ProductVariant
		var totalStock int = 0

		for i := range p.Variants {
			v := &p.Variants[i]
			totalStock += v.Stock
			if v.IsDefault {
				defaultVariant = v
			}
		}

		if defaultVariant == nil {
			defaultVariant = &p.Variants[0]
		}

		currentPrice := defaultVariant.Price
		if defaultVariant.SalePrice > 0 && defaultVariant.SalePrice < defaultVariant.Price {
			currentPrice = defaultVariant.SalePrice
		}

		p.Price = currentPrice
		p.Stock = totalStock

		var firstImage string = ""
		for _, img := range defaultVariant.Images {
			if img.IsPrimary {
				if img.Image != "" {
					firstImage = img.Image
				} else {
					firstImage = img.ImageURL
				}
				break
			}
		}
		if firstImage == "" && len(defaultVariant.Images) > 0 {
			if defaultVariant.Images[0].Image != "" {
				firstImage = defaultVariant.Images[0].Image
			} else {
				firstImage = defaultVariant.Images[0].ImageURL
			}
		}

		if firstImage != "" {
			p.ImageURL = firstImage
		}
	}
	return nil
}

func (v *ProductVariant) BeforeSave(tx *gorm.DB) (err error) {
	if v.ID != 0 {
		var stats struct {
			AvgRating float64
			Count     int64
		}
		err := tx.Model(&VariantReview{}).
			Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as count").
			Where("variant_id = ?", v.ID).
			Scan(&stats).Error
		if err != nil {
			return err
		}
		v.AverageRating = stats.AvgRating
		v.ReviewCount = int(stats.Count)
	} else {
		v.AverageRating = 0
		v.ReviewCount = 0
	}
	return nil
}

type ProductRepository interface {
	Create(product *Product) error
	GetAll(search string, categoryID uint, subCategoryID uint) ([]Product, error)
	GetByID(id uint) (*Product, error)
	Update(product *Product) error
	Delete(id uint) error

	// Variant CRUD
	CreateVariant(variant *ProductVariant) error
	UpdateVariant(variant *ProductVariant) error
	DeleteVariant(id uint) error
	GetVariantByID(id uint) (*ProductVariant, error)

	// Image management
	AddVariantImage(img *VariantImage) error
	DeleteVariantImage(id uint) error

	// Review management
	AddVariantReview(rev *VariantReview) error
	GetVariantReviews(variantID uint) ([]VariantReview, error)
}

type ProductUseCase interface {
	GetAllProducts(search string, categoryID uint, subCategoryID uint) ([]Product, error)
	GetProduct(id uint) (*Product, error)
	CreateProduct(product *Product) error
	UpdateProduct(product *Product) error
	DeleteProduct(id uint) error

	// Variant UseCase
	CreateVariant(variant *ProductVariant) error
	UpdateVariant(variant *ProductVariant) error
	DeleteVariant(id uint) error
	GetVariantByID(id uint) (*ProductVariant, error)

	// Image UseCase
	AddVariantImage(img *VariantImage) error
	DeleteVariantImage(id uint) error

	// Review UseCase
	AddVariantReview(rev *VariantReview) error
	GetVariantReviews(variantID uint) ([]VariantReview, error)
}

