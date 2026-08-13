package main

import (
	"log"
	"os"
	"time"
	"fmt"
	"math/rand"
	"trendybackend/internal/api/handler"
	"trendybackend/internal/api/middleware"
	"trendybackend/internal/api/websocket"
	"trendybackend/internal/domain"
	"trendybackend/internal/repository"
	"trendybackend/internal/usecase"
	"trendybackend/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if available (ignored in cloud production environment like Railway)
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")
	
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "super-secret-key-123")
	}

	// Initialize Database
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	log.Println("Migrating database models...")
	err = database.AutoMigrate(
		&domain.User{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Tag{},
		&domain.Product{},
		&domain.ProductVariant{},
		&domain.VariantImage{},
		&domain.VariantReview{},
		&domain.Promotion{},
		&domain.Order{},
		&domain.ZohoProduct{},
		&domain.ZohoProductReview{},
		&domain.Setting{},
		&domain.ProductActivity{},
		&domain.ProductView{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Recalculate variant average ratings and review counts on startup to fix any stale cached counts
	log.Println("Recalculating variant ratings and review counts...")
	var variants []domain.ProductVariant
	if err := database.Find(&variants).Error; err == nil {
		log.Printf("Found %d variants to recalculate\n", len(variants))
		for _, v := range variants {
			var stats struct {
				AvgRating float64
				Count     int64
			}
			database.Model(&domain.VariantReview{}).
				Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(id) as count").
				Where("variant_id = ?", v.ID).
				Scan(&stats)

			log.Printf("Variant %d: AvgRating=%f, Count=%d\n", v.ID, stats.AvgRating, stats.Count)

			database.Model(&domain.ProductVariant{}).
				Where("id = ?", v.ID).
				Updates(map[string]interface{}{
					"average_rating": stats.AvgRating,
					"review_count":   int(stats.Count),
				})
		}
	} else {
		log.Printf("Error finding variants: %v\n", err)
	}

	// Seed product sales for realistic dashboard analytics representation
	var salesCount int64
	database.Model(&domain.Product{}).Where("sales > 0").Count(&salesCount)
	if salesCount == 0 {
		log.Println("Seeding product sales counts for dashboard representation...")
		var products []domain.Product
		if err := database.Find(&products).Error; err == nil {
			salesAmounts := []int{142, 98, 76, 54, 32, 21, 15, 12, 8}
			for i, p := range products {
				sales := 5
				if i < len(salesAmounts) {
					sales = salesAmounts[i]
				} else {
					sales = 5 + (int(p.ID) % 15)
				}
				database.Model(&p).Update("sales", sales)
			}
		}
	}

	// Seed recommended Zoho products if none are recommended
	var recommendedCount int64
	database.Model(&domain.ZohoProduct{}).Where("is_recommended = ?", true).Count(&recommendedCount)
	if recommendedCount == 0 {
		log.Println("Seeding recommended Zoho products...")
		var zohoProducts []domain.ZohoProduct
		if err := database.Limit(6).Find(&zohoProducts).Error; err == nil {
			for _, p := range zohoProducts {
				database.Model(&p).Update("is_recommended", true)
			}
			log.Printf("Marked %d Zoho products as recommended\n", len(zohoProducts))
		}
	}

	// Seed trending Zoho products if less than 12 are trending
	var trendingCount int64
	database.Model(&domain.ZohoProduct{}).Where("is_trending = ?", true).Count(&trendingCount)
	if trendingCount < 12 {
		log.Println("Seeding trending Zoho products...")
		var zohoProducts []domain.ZohoProduct
		// Fetch up to 20 products and mark them as trending
		if err := database.Limit(20).Find(&zohoProducts).Error; err == nil {
			for _, p := range zohoProducts {
				database.Model(&p).Update("is_trending", true)
			}
			log.Printf("Marked %d Zoho products as trending\n", len(zohoProducts))
		}
	}

	// Seed sales volume for Zoho products if they have sales_volume <= 0
	var zohoProductsToSeed []domain.ZohoProduct
	if err := database.Where("sales_volume <= 0").Find(&zohoProductsToSeed).Error; err == nil && len(zohoProductsToSeed) > 0 {
		log.Println("Seeding sales volume for Zoho products with 0 sales...")
		salesAmounts := []int{142, 98, 76, 54, 45, 32, 21, 15, 12, 8}
		for i, p := range zohoProductsToSeed {
			sales := 10
			if i < len(salesAmounts) {
				sales = salesAmounts[i]
			} else {
				sales = 10 + (int(p.CategoryID*3+p.SubCategoryID*7)%35 + (i % 7))
			}
			database.Model(&p).Update("sales_volume", sales)
		}
		log.Printf("Seeded sales volume for %d Zoho products\n", len(zohoProductsToSeed))
	}

	// Seed mock orders if there are less than 5 orders in the database
	var orderCount int64
	database.Model(&domain.Order{}).Count(&orderCount)
	if orderCount < 5 {
		log.Println("Seeding mock orders for store statistics...")
		names := []string{"Junaid Rashid", "Ahmed Al-Farsi", "Sarah Connor", "John Doe", "Jane Smith", "Michael Scott", "Jim Halpert", "Pam Beesly"}
		emails := []string{"junaid@example.com", "ahmed@example.com", "sarah@example.com", "john@example.com", "jane@example.com", "michael@example.com", "jim@example.com", "pam@example.com"}
		amounts := []float64{120.50, 45.00, 250.00, 85.20, 15.00, 110.00, 65.40, 95.00}
		statuses := []string{"Delivered", "Delivered", "Preparing", "Delivered", "Cancelled", "Delivered", "OutForDelivery", "Delivered"}

		var topZohoProducts []domain.ZohoProduct
		database.Limit(8).Find(&topZohoProducts)

		for i := 0; i < len(names); i++ {
			orderDate := time.Now().AddDate(0, 0, -i)
			itemsJson := `[{"name":"Mock Product","quantity":1,"price":` + fmt.Sprintf("%.2f", amounts[i]) + `,"imageUrl":""}]`

			if len(topZohoProducts) > i {
				zp := topZohoProducts[i]
				qty := 2 + (i % 5)
				itemsJson = fmt.Sprintf(`[{"zoho_item_id":"%s","name":%q,"quantity":%d,"price":%.2f,"imageUrl":%q}]`,
					zp.ZohoItemID, zp.Name, qty, zp.Rate, zp.CustomImage)
			}

			mockOrder := domain.Order{
				CustomerName:    names[i],
				CustomerEmail:   emails[i],
				CustomerPhone:   "+966501234567",
				CustomerAddress: "Riyadh, Saudi Arabia",
				PaymentMethod:   "Credit Card",
				ItemsJson:       itemsJson,
				TotalAmount:     amounts[i],
				Status:          statuses[i],
				CreatedAt:       orderDate,
				UpdatedAt:       orderDate,
			}
			database.Create(&mockOrder)
		}
		log.Println("Successfully seeded 8 mock orders with multi-channel product items")
	}

	// Seed mock product views if there are no view records in the database
	var productViewCount int64
	database.Model(&domain.ProductView{}).Count(&productViewCount)
	if productViewCount == 0 {
		log.Println("Seeding mock product views for Top Viewed Products section...")
		var zohoProducts []domain.ZohoProduct
		if err := database.Limit(10).Find(&zohoProducts).Error; err == nil {
			for idx, p := range zohoProducts {
				viewsToSeed := 15 - idx
				for v := 0; v < viewsToSeed; v++ {
					viewDate := time.Now().Add(-time.Duration(v*2) * time.Hour)
					mockView := domain.ProductView{
						ProductID: p.ZohoItemID,
						GuestID:   fmt.Sprintf("guest_seed_user_%d", v),
						CreatedAt: viewDate,
					}
					database.Create(&mockView)
				}
			}
			log.Println("Successfully seeded mock product views")
		}
	}

	// Seed mock reviews for 80% of Zoho products if the review table is empty
	var reviewsCount int64
	database.Model(&domain.ZohoProductReview{}).Count(&reviewsCount)
	if reviewsCount < 100 {
		log.Println("Seeding mock reviews for 80% of Zoho Products...")
		var allProducts []domain.ZohoProduct
		if err := database.Find(&allProducts).Error; err == nil && len(allProducts) > 0 {
			englishNames := []string{
				"Michael Smith", "David Miller", "Emma Watson", "Sophia Johnson",
				"James Williams", "Olivia Brown", "Alexander Davis", "Isabella Martinez",
				"William Anderson", "Emily Taylor", "Liam Thomas", "Mia Moore",
			}
			arabicNames := []string{
				"أحمد القحطاني", "سارة الشمري", "خالد الحربي", "فاطمة العتيبي",
				"عبدالرحمن الدوسري", "نورة البقمي", "محمد العنزي", "منى المطيري",
				"يوسف الغامدي", "ريم الفهد", "سلطان العتيبي", "حصة السديري",
			}
			englishReviews := []string{
				"Absolutely love this! The quality is amazing and it arrived so fast.",
				"Very satisfied with the purchase. Matches the description perfectly.",
				"Good quality product, highly recommended!",
				"Excellent value for money. Will definitely buy again.",
				"Great customer service and the product is exactly what I wanted.",
				"Wonderful experience, item packaging was very neat and safe.",
				"Super fast delivery and fantastic product build quality.",
				"Perfect fit and very comfortable. Five stars!",
			}
			arabicReviews := []string{
				"المنتج رائع جداً وجودته ممتازة، أنصح بشدّة بشرائه.",
				"توصيل سريع والمنتج مطابق للوصف تماماً. شكراً لكم.",
				"جودة عالية وسعر مناسب جداً مقارنة بالسوق.",
				"تجربة شراء ممتازة وتغليف نظيف وآمن للغاية.",
				"منتج ممتاز وتوصيل أسرع مما توقعت. يستحق ٥ نجوم.",
				"رائع جداً ومريح في الاستخدام، سأطلب المزيد بالتأكيد.",
				"الخامة ممتازة ومطابقة للصورة تماماً. تعامل راقي وسريع.",
				"أفضل منتج جربته في هذه الفئة، جودة تفوق التوقعات.",
			}

			targetCount := int(float64(len(allProducts)) * 0.8)
			var mockReviews []domain.ZohoProductReview

			r := rand.New(rand.NewSource(42))

			for i := 0; i < targetCount; i++ {
				p := allProducts[i]
				numReviews := r.Intn(3) + 1
				for j := 0; j < numReviews; j++ {
					isArabic := r.Intn(2) == 0
					var name, reviewText string
					if isArabic {
						name = arabicNames[r.Intn(len(arabicNames))]
						reviewText = arabicReviews[r.Intn(len(arabicReviews))]
					} else {
						name = englishNames[r.Intn(len(englishNames))]
						reviewText = englishReviews[r.Intn(len(englishReviews))]
					}
					
					rating := r.Intn(2) + 4 // 4 or 5 stars
					
					mockReview := domain.ZohoProductReview{
						ZohoProductID: p.ZohoItemID,
						CustomerID:    uint(r.Intn(1000) + 1),
						CustomerName:  name,
						Rating:        rating,
						Review:        reviewText,
						CreatedAt:     time.Now().Add(-time.Duration(r.Intn(30)) * 24 * time.Hour),
					}
					mockReviews = append(mockReviews, mockReview)
				}
			}

			if len(mockReviews) > 0 {
				if err := database.CreateInBatches(&mockReviews, 1000).Error; err != nil {
					log.Printf("Failed to batch seed reviews: %v\n", err)
				} else {
					log.Printf("Successfully seeded %d mock product reviews for %d products\n", len(mockReviews), targetCount)
				}
			}
		}
	}

	// Initialize Dependencies
	productRepo := repository.NewProductRepository(database)
	productUseCase := usecase.NewProductUseCase(productRepo)
	productHandler := handler.NewProductHandler(productUseCase)

	categoryRepo := repository.NewCategoryRepository(database)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	subCategoryRepo := repository.NewSubCategoryRepository(database)
	subCategoryUseCase := usecase.NewSubCategoryUseCase(subCategoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryUseCase, subCategoryUseCase)

	promotionRepo := repository.NewPromotionRepository(database)
	promotionUseCase := usecase.NewPromotionUseCase(promotionRepo)
	promotionHandler := handler.NewPromotionHandler(promotionUseCase)

	orderRepo := repository.NewOrderRepository(database)
	orderUseCase := usecase.NewOrderUseCase(orderRepo)
	orderHandler := handler.NewOrderHandler(orderUseCase)

	userRepo := repository.NewUserRepository(database)
	authUseCase := usecase.NewAuthUsecase(userRepo)
	authHandler := handler.NewAuthHandler(authUseCase)

	analyticsUseCase := usecase.NewAnalyticsUseCase(database)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsUseCase)

	notificationUseCase := usecase.NewNotificationUseCase(database)
	notificationHandler := handler.NewNotificationHandler(notificationUseCase)

	tagRepo := repository.NewTagRepository(database)
	tagUseCase := usecase.NewTagUseCase(tagRepo)
	tagHandler := handler.NewTagHandler(tagUseCase)

	geminiHandler := handler.NewGeminiHandler()
	zohoHandler := handler.NewZohoHandler(database)
	settingHandler := handler.NewSettingHandler(database)
	uiHandler := handler.NewUIHandler(database)

	// Seed Super Admin
	if err := authUseCase.SeedSuperAdmin(); err != nil {
		log.Printf("Warning: failed to seed super admin: %v", err)
	}

	// Ensure uploads directory exists
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		// Detect if we are in cmd/api and project uploads is at ../../uploads
		if _, err := os.Stat("../../uploads"); err == nil {
			uploadsDir = "../../uploads"
		} else {
			uploadsDir = "./uploads"
		}
		os.Setenv("UPLOADS_DIR", uploadsDir)
	}
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	log.Printf("Serving uploads from: %s", uploadsDir)

	// Setup Gin router
	r := gin.Default()

	// CORS Middleware — must be registered before Static and API routes
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		// Allow cross-origin image loading (required for Flutter Web NetworkImage)
		c.Writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	r.Static("/uploads", uploadsDir)

	// Routes
	api := r.Group("/api/v1")
	{
		api.POST("/upload", productHandler.UploadImage)

		promotions := api.Group("/promotions")
		{
			promotions.GET("", promotionHandler.GetAll)
			promotions.GET("/:id", promotionHandler.GetByID)
			promotions.POST("", promotionHandler.Create)
			promotions.PUT("/:id", promotionHandler.Update)
			promotions.DELETE("/:id", promotionHandler.Delete)
		}
		orders := api.Group("/orders")
		{
			orders.GET("", orderHandler.GetAll)
			orders.GET("/:id", orderHandler.GetByID)
			orders.POST("", orderHandler.Create)
			orders.PUT("/:id", orderHandler.Update)
			orders.DELETE("/:id", orderHandler.Delete)
		}
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetAllCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.POST("", categoryHandler.CreateCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}
		subCategories := api.Group("/subcategories")
		{
			subCategories.GET("", categoryHandler.GetAllSubCategories)
			subCategories.GET("/:id", categoryHandler.GetSubCategory)
			subCategories.POST("", categoryHandler.CreateSubCategory)
			subCategories.PUT("/:id", categoryHandler.UpdateSubCategory)
			subCategories.DELETE("/:id", categoryHandler.DeleteSubCategory)
		}
		tags := api.Group("/tags")
		{
			tags.GET("", tagHandler.GetAll)
			tags.GET("/:id", tagHandler.GetByID)
			tags.POST("", tagHandler.Create)
			tags.PUT("/:id", tagHandler.Update)
			tags.DELETE("/:id", tagHandler.Delete)
		}

		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.PUT("/profile", authHandler.UpdateProfile)
		}

		api.GET("/customers", authHandler.GetCustomers)
		api.DELETE("/customers/:id", authHandler.DeleteCustomer)

		// Analytics Routes
		api.GET("/analytics/dashboard", analyticsHandler.GetDashboardStats)

		// Notification Routes
		api.GET("/notifications", notificationHandler.GetNotifications)

		// Settings Routes
		api.GET("/settings/:key", settingHandler.GetSetting)
		api.POST("/settings", settingHandler.SaveSetting)

		// UI Management Routes
		api.POST("/ui/upload-video", uiHandler.UploadHeroVideo)
		api.DELETE("/ui/delete-video", uiHandler.DeleteHeroVideo)

		// WebSocket Route
		api.GET("/ws", func(c *gin.Context) {
			websocket.GetHub().HandleConnection(c)
		})

		// Gemini Routes
		api.GET("/gemini/generate-description", geminiHandler.GenerateDescription)
		api.GET("/gemini/generate-seo", geminiHandler.GenerateSEO)
		api.GET("/gemini/pricing-insight", geminiHandler.GeneratePricingInsight)
		api.POST("/gemini/generate-photoshoot", geminiHandler.GeneratePhotoshoot)
		api.POST("/gemini/scan-product", geminiHandler.ScanProduct)
		api.GET("/gemini/zoho-ai-assist", geminiHandler.GenerateZohoAIAssist)

		// Zoho Routes
		api.POST("/zoho/webhook", zohoHandler.ReceiveWebhook)
		api.GET("/zoho/products", zohoHandler.GetProducts)
		api.GET("/zoho/products/:id", zohoHandler.GetProduct)
		api.GET("/zoho/brands", zohoHandler.GetBrands)
		api.POST("/zoho/products", zohoHandler.CreateProduct)
		api.PUT("/zoho/products/:id/visibility", zohoHandler.ToggleVisibility)
		api.PUT("/zoho/products/:id/recommended", zohoHandler.ToggleRecommended)
		api.PUT("/zoho/products/:id", zohoHandler.UpdateProduct)
		api.GET("/zoho/products/:id/image", zohoHandler.GetProductImage)
		api.POST("/zoho/products/:id/reviews", zohoHandler.CreateReview)
		api.POST("/zoho/products/:id/activity", zohoHandler.LogProductActivity)
		api.POST("/zoho/products/:id/view", zohoHandler.LogProductView)
		api.GET("/zoho/products/most-viewed", zohoHandler.GetMostViewedProducts)
		api.GET("/zoho/products/ice-cream-fest", zohoHandler.GetIceCreamFestProducts)
		api.GET("/zoho/products/baking-fest", zohoHandler.GetBakingFestProducts)
		api.GET("/zoho/products/:id/reviews", zohoHandler.GetReviews)
		api.POST("/zoho/sync", zohoHandler.SyncProducts)
		api.POST("/zoho/sync-trending", zohoHandler.SyncTrendingProducts)

		// Admin Management (Super Admin Only)
		admins := api.Group("/admins")
		admins.Use(middleware.AuthMiddleware())
		{
			admins.GET("", authHandler.GetAdmins)
			admins.POST("", authHandler.CreateAdmin)
			admins.DELETE("/:id", authHandler.DeleteAdmin)
		}
	}

	// Basic health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Get port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
