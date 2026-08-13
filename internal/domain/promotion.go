package domain

type Promotion struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Title              string    `json:"title"`
	Subtitle           string    `json:"subtitle"`
	ImageUrl           string    `json:"image_url"`
	DiscountPercentage float64   `json:"discount_percentage"`
	IsActive           bool      `json:"is_active" gorm:"default:true"`
	Type               string    `json:"type" gorm:"default:'carousel'"`
	Products           []ZohoProduct `json:"products" gorm:"many2many:promotion_zoho_products;joinForeignKey:promotion_id;joinReferences:zoho_item_id"`
	ProductIDs         []string      `json:"product_ids" gorm:"-"`
}

type PromotionRepository interface {
	Create(promotion *Promotion) error
	GetAll() ([]Promotion, error)
	GetByID(id uint) (*Promotion, error)
	Update(promotion *Promotion) error
	Delete(id uint) error
}

type PromotionUseCase interface {
	CreatePromotion(promotion *Promotion) error
	GetAllPromotions() ([]Promotion, error)
	GetPromotion(id uint) (*Promotion, error)
	UpdatePromotion(promotion *Promotion) error
	DeletePromotion(id uint) error
}
