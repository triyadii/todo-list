package utils

import (
	"errors"
	"strconv"
	"time"

	"todo-backend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Timezone string    `json:"timezone"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, email, timezone string) (string, error) {
	expHours, err := strconv.Atoi(config.AppConfig.JWTExpireHr)
	if err != nil || expHours <= 0 {
		expHours = 72
	}

	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		Timezone: timezone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}
