package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	DB *gorm.DB
}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{
		DB: database.GetDB(),
	}
}

// ============================
// POST Activity + Upload Dokumen
// ============================
func (h *ActivityHandler) StoreActivity(c *gin.Context) {

	tipeActive := c.PostForm("tipe_active")
	judulActive := c.PostForm("judul_active")
	timeActive := c.PostForm("time") // format: HH:mm
	userIDStr := c.PostForm("user_id")

	if tipeActive == "" || judulActive == "" || timeActive == "" || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Semua field wajib diisi",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "user_id tidak valid",
		})
		return
	}

	// ============================
	// Upload Dokumen (Optional)
	// ============================
	var dokumenURL *string

	file, err := c.FormFile("dokumen")
	if err == nil && file != nil {
		uploadDir := "uploads/activity"
		os.MkdirAll(uploadDir, os.ModePerm)

		filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(file.Filename)
		filePath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Gagal upload dokumen",
			})
			return
		}

		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		url := scheme + "://" + c.Request.Host + "/" + filePath
		dokumenURL = &url
	}

	today := time.Now().Format("2006-01-02")
	fullTime := today + " " + timeActive

	// ============================
	// Simpan ke Database
	// ============================
	query := `
		INSERT INTO activity 
		(tipe_active, judul_active, time, dokumen_link, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	if err := h.DB.Exec(
		query,
		tipeActive,
		judulActive, // ⬅️ ini yang benar
		fullTime,    // ⬅️ timestamp lengkap
		dokumenURL,
		userID,
		time.Now(),
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"message":       "Gagal menyimpan activity",
			"error_details": err.Error(),
		})
		return
	}

	// ============================
	// Response
	// ============================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Activity berhasil disimpan",
		"data": gin.H{
			"tipe_active":  tipeActive,
			"judul_active": judulActive,
			"time":         timeActive,
			"dokumen_link": dokumenURL,
			"user_id":      userID,
		},
	})
}
