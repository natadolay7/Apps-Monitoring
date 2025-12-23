package handlers

import (
	"net/http"
	"strconv"

	"api_patroliku_docker/database"
	"api_patroliku_docker/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		DB: database.GetDB(),
	}
}

// GetAllUsers mengambil semua data users
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var users []models.User

	// Query dengan pagination sederhana
	limit := 50
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// Eksekusi query
	result := h.DB.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data users",
			"error":   result.Error.Error(),
		})
		return
	}

	// Hitung total
	var total int64
	h.DB.Model(&models.User{}).Count(&total)

	// Format response tanpa password
	var response []gin.H
	for _, user := range users {
		response = append(response, gin.H{
			"id": user.ID,

			"email":      user.Email,
			"name":       user.Name,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data users berhasil diambil",
		"data":    response,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": (int(total) + limit - 1) / limit,
		},
	})
}

// GetUserByID mengambil user berdasarkan ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID harus berupa angka",
		})
		return
	}

	var user models.User
	result := h.DB.First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data user berhasil diambil",
		"data": gin.H{
			"id": user.ID,

			"email":      user.Email,
			"name":       user.Name,
			"updated_at": user.UpdatedAt,
		},
	})
}

// CreateUser membuat user baru
func (h *UserHandler) CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}

	// Bind dan validasi input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Validasi gagal",
			"errors":  err.Error(),
		})
		return
	}

	// Cek apakah email sudah terdaftar
	var existingUser models.User
	if err := h.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "Email sudah terdaftar",
		})
		return
	}

	// Buat user baru
	user := models.User{

		Email:    input.Email,
		Password: input.Password, // Dalam real app, harus di-hash!
		Name:     input.FullName,
	}

	result := h.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal membuat user",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User berhasil dibuat",
		"data": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"name":       user.Name,
			"created_at": user.CreatedAt,
		},
	})
}
