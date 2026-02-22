package utils

import (
	"errors"
	"os"
	"time"

	"github.com/dr15/internship-hub-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID      uuid.UUID       `json:"userId"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	UnitKerjaID *uuid.UUID      `json:"unitKerjaId,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		UnitKerjaID: user.UnitKerjaID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
