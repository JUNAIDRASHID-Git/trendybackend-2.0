package usecase

import (
	"errors"
	"os"
	"time"
	"trendybackend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo domain.UserRepository
}

func NewAuthUsecase(userRepo domain.UserRepository) *AuthUsecase {
	return &AuthUsecase{userRepo: userRepo}
}

func (u *AuthUsecase) Login(email, password string) (string, *domain.User, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (u *AuthUsecase) SeedSuperAdmin() error {
	email := "junaidrashid678@gmail.com"
	password := "Pazhom123#"
	
	_, err := u.userRepo.FindByEmail(email)
	if err == nil {
		return nil // Already exists
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	superAdmin := &domain.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: "Super",
		LastName:  "Admin",
		Role:      "super_admin",
	}

	return u.userRepo.Create(superAdmin)
}

func (u *AuthUsecase) GetAllAdmins() ([]domain.User, error) {
	return u.userRepo.FindAll()
}

func (u *AuthUsecase) CreateAdmin(email, password, firstName, lastName string) error {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	admin := &domain.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Role:      "admin",
	}

	return u.userRepo.Create(admin)
}

func (u *AuthUsecase) DeleteAdmin(id uint) error {
	return u.userRepo.Delete(id)
}

func (u *AuthUsecase) GetUserByEmail(email string) (*domain.User, error) {
	return u.userRepo.FindByEmail(email)
}

func (u *AuthUsecase) CreateCustomer(email, password, firstName, lastName, phone string) (*domain.User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	customer := &domain.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Role:      "customer",
	}

	err := u.userRepo.Create(customer)
	return customer, err
}

func (u *AuthUsecase) UpdateUser(user *domain.User) error {
	return u.userRepo.Update(user)
}

func (u *AuthUsecase) GetCustomersWithStats() ([]domain.CustomerWithStats, error) {
	return u.userRepo.GetCustomersWithStats()
}

func (u *AuthUsecase) GetUserByID(id uint) (*domain.User, error) {
	return u.userRepo.FindByID(id)
}

func (u *AuthUsecase) DeleteCustomer(id uint) error {
	return u.userRepo.Delete(id)
}
