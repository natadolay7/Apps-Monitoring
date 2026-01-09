package config

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JWTSecret = []byte("patroliku-api-2025-token") // Ganti dengan secret yang aman

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	BranchID int    `json:"branch_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, email, userType string, branchID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token berlaku 24 jam

	// expirationTime := time.Now().Add(1 * time.Minute) // Token berlaku 1 menit

	claims := &JWTClaims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		BranchID: branchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "api-patroliku",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
