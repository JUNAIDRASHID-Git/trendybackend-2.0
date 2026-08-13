package handler

import (
	"net/http"
	"strconv"
	"trendybackend/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authUsecase.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) GetAdmins(c *gin.Context) {
	// Role check should be in middleware, but we can double check here
	role, _ := c.Get("role")
	if role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can manage admins"})
		return
	}

	admins, err := h.authUsecase.GetAllAdmins()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admins)
}

func (h *AuthHandler) CreateAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can manage admins"})
		return
	}

	var input struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authUsecase.CreateAdmin(input.Email, input.Password, input.FirstName, input.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "admin created successfully"})
}

func (h *AuthHandler) DeleteAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only super admin can manage admins"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	err := h.authUsecase.DeleteAdmin(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "admin deleted successfully"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if already exists
	if existingUser, err := h.authUsecase.GetUserByEmail(input.Email); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "user already registered",
			"status":  "existing",
			"user":    existingUser,
		})
		return
	}

	password := input.Password
	if password == "" {
		password = "oauth-placeholder-password-123!"
	}

	customer, err := h.authUsecase.CreateCustomer(input.Email, password, input.FirstName, input.LastName, input.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"status":  "new",
		"user":    customer,
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var input struct {
		Email           string   `json:"email" binding:"required"`
		Phone           *string  `json:"phone"`
		AddressCity     *string  `json:"address_city"`
		AddressDistrict *string  `json:"address_district"`
		AddressLandmark *string  `json:"address_landmark"`
		AddressLat      *float64 `json:"address_lat"`
		AddressLon      *float64 `json:"address_lon"`
		CartItems       *string  `json:"cart_items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authUsecase.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	if input.AddressCity != nil {
		user.AddressCity = *input.AddressCity
	}
	if input.AddressDistrict != nil {
		user.AddressDistrict = *input.AddressDistrict
	}
	if input.AddressLandmark != nil {
		user.AddressLandmark = *input.AddressLandmark
	}
	if input.AddressLat != nil {
		user.AddressLat = *input.AddressLat
	}
	if input.AddressLon != nil {
		user.AddressLon = *input.AddressLon
	}
	if input.CartItems != nil {
		user.CartItems = *input.CartItems
	}

	err = h.authUsecase.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully", "user": user})
}

func (h *AuthHandler) GetCustomers(c *gin.Context) {
	customers, err := h.authUsecase.GetCustomersWithStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

func (h *AuthHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	user, err := h.authUsecase.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if user.Role != "customer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only customer accounts can be deleted"})
		return
	}

	err = h.authUsecase.DeleteCustomer(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}
