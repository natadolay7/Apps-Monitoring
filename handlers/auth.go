package handlers

import (
	"fmt"
	"net/http"
	"time"

	"api_patroliku_docker/config"
	"api_patroliku_docker/database"
	"api_patroliku_docker/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		DB: database.GetDB(),
	}
}

// Login godoc
// @Summary Login user
// @Description Login user dengan email dan password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest

	// Bind request
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// Validasi input
	if loginReq.Username == "" || loginReq.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Username dan password wajib diisi",
		})
		return
	}

	// Query untuk mendapatkan user
	var user models.UserLogin
	query := `
        SELECT
            u.id,
            u."name",
            u.email,
            u.password,
            b.id AS branch_id,
            ut."name" AS user_type,
            b."name" AS branch_name,
            u.position_id,
            u.profile_photo_path,
            u.fcm_token
        FROM users u
        INNER JOIN user_type ut ON ut.id = u.user_type_id
        INNER JOIN user_tad_information uti ON u.id = uti.user_id
        INNER JOIN branch b ON b.id = uti.branch_id
        WHERE u.email = ? AND u.deleted_at IS NULL
        LIMIT 1
    `

	// Gunakan Username sebagai parameter untuk query di field email
	if err := h.DB.Raw(query, loginReq.Username).Scan(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
		return
	}

	// Cek jika user ditemukan
	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Username atau password salah",
		})
		return
	}

	// Verifikasi password
	// Asumsi password di database sudah di-hash dengan bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		// Jika password masih plain text (untuk migration)
		if user.Password != loginReq.Password {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Username atau password salah",
			})
			return
		}
		// Log warning untuk migration
		fmt.Println("WARNING: Password masih plain text untuk user:", user.Email)
	}

	// Generate JWT token
	token, err := config.GenerateToken(user.ID, user.Email, user.UserType, user.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal generate token",
			"error":   err.Error(),
		})
		return
	}

	// Update FCM token jika ada di request
	if loginReq.FCMToken != "" && loginReq.FCMToken != user.FCMToken {
		h.DB.Model(&models.User{}).Where("id = ?", user.ID).
			Update("fcm_token", loginReq.FCMToken)
	}

	// Buat response TANPA password
	response := models.LoginResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email, // Ini akan berisi username (peg261295)
		BranchID:         user.BranchID,
		UserType:         user.UserType,
		BranchName:       user.BranchName,
		PositionID:       user.PositionID,
		ProfilePhotoPath: user.ProfilePhotoPath,
		FCMToken:         user.FCMToken,
		Token:            token,
		TokenType:        "Bearer",
		ExpiresIn:        time.Now().Add(24 * time.Hour).Unix(),
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil",
		"data":    response,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user dan hapus FCM token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Dapatkan user ID dari context (setelah middleware auth)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	// Hapus FCM token
	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("fcm_token", "").Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal logout",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout berhasil",
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user profile from token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Dapatkan user ID dari context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	// Query untuk mendapatkan user profile
	var user models.LoginResponse
	query := `
        SELECT
            u.id,
            u."name",
            u.email,
            b.id AS branch_id,
            ut."name" AS user_type,
            b."name" AS branch_name,
            u.position_id,
            u.profile_photo_path,
            u.fcm_token
        FROM users u
        INNER JOIN user_type ut ON ut.id = u.user_type_id
        INNER JOIN user_tad_information uti ON u.id = uti.user_id
        INNER JOIN branch b ON b.id = uti.branch_id
        WHERE u.id = ? AND u.deleted_at IS NULL
        LIMIT 1
    `

	if err := h.DB.Raw(query, userID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data profile",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile berhasil diambil",
		"data":    user,
	})
}
