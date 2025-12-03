package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"Zjmf-kvm/internal/model"
)

var ErrUserExists = errors.New("User already exists")

func RegisterUser(db *gorm.DB, email, password string) (*model.User, error) {
	var existing model.User
	if err := db.Where("Email = ?", email).First(&existing).Error; err == nil {
		return nil, ErrUserExists
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{Email: email, Password: string(hashed), Role: "user"}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GenerateToken(jwtSecret string, user *model.User, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"exp":  time.Now().Add(ttl).Unix(),
		"role": user.Role,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(jwtSecret))
}
