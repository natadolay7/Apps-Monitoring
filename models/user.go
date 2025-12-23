package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"size:255;uniqueIndex;not null"`
	Password string `gorm:"size:255;not null"`
	Name     string `gorm:"size:255"`
}

// LoginRequest untuk request login
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // Ubah dari Email ke Username
	Password string `json:"password" binding:"required"`
	FCMToken string `json:"fcm_token,omitempty"`
}

// LoginResponse untuk response login
type LoginResponse struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	BranchID         int    `json:"branch_id"`
	UserType         string `json:"user_type"`
	BranchName       string `json:"branch_name"`
	PositionID       int    `json:"position_id"`
	ProfilePhotoPath string `json:"profile_photo_path"`
	FCMToken         string `json:"fcm_token"`
	Token            string `json:"token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
}

type UserLogin struct {
	ID               int    `gorm:"column:id"`
	Name             string `gorm:"column:name"`
	Email            string `gorm:"column:email"`
	Password         string `gorm:"column:password"` // Tambahkan ini
	BranchID         int    `gorm:"column:branch_id"`
	UserType         string `gorm:"column:user_type"`
	BranchName       string `gorm:"column:branch_name"`
	PositionID       int    `gorm:"column:position_id"`
	ProfilePhotoPath string `gorm:"column:profile_photo_path"`
	FCMToken         string `gorm:"column:fcm_token"`
}
