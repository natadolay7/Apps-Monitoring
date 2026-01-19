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

	// Ambil user_id dari JWT middleware
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User tidak terautentikasi",
		})
		return
	}
	userID := userIDRaw.(int)

	// Ambil form data
	tipeActive := c.PostForm("tipe_active")
	judulActive := c.PostForm("judul_active")
	deskripsi := c.PostForm("deskripsi")
	timeActive := c.PostForm("time") // format: HH:mm:ss

	if tipeActive == "" || judulActive == "" || timeActive == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "tipe_active, judul_active, dan time wajib diisi",
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

		// scheme := "http"
		// if c.Request.TLS != nil {
		// 	scheme = "https"
		// }

		// url := scheme + "://" + c.Request.Host + "/" + filePath
		relativePath := "/" + filePath
		dokumenURL = &relativePath
	}

	// Gabungkan tanggal hari ini + jam dari input
	today := time.Now().Format("2006-01-02")
	fullTime := today + " " + timeActive

	// ============================
	// Simpan ke Database
	// ============================
	query := `
		INSERT INTO activity 
		(tipe_active, judul_active, deskripsi, time, dokumen_link, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	if err := h.DB.Exec(
		query,
		tipeActive,
		judulActive,
		deskripsi,
		fullTime,
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
			"deskripsi":    deskripsi,
			"time":         fullTime,
			"dokumen_link": dokumenURL,
			"user_id":      userID,
		},
	})
}

func (h *ActivityHandler) GetActivityByUser(c *gin.Context) {

	// Ambil user_id dari JWT middleware
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User tidak terautentikasi",
		})
		return
	}

	userID := userIDRaw.(int)

	// Ambil filter tanggal (optional)
	date := c.Query("date") // format: YYYY-MM-DD

	type ActivityResponse struct {
		ID          int     `json:"id"`
		TipeActive  string  `json:"tipe_active"`
		JudulActive string  `json:"judul_active"`
		Time        string  `json:"time"`
		DokumenLink *string `json:"dokumen_link"`
		Deskripsi   *string `json:"deskripsi"`
		UserID      int     `json:"user_id"`
		CreatedAt   string  `json:"created_at"`
	}

	var activities []ActivityResponse

	query := `
		SELECT 
			id, 
			tipe_active, 
			judul_active, 
			time, 
			dokumen_link, 
			user_id, 
			deskripsi,
			created_at
		FROM activity
		WHERE user_id = ?
	`

	params := []interface{}{userID}

	if date != "" {
		query += " AND DATE(created_at) = ?"
		params = append(params, date)
	}

	query += " ORDER BY created_at DESC"

	if err := h.DB.Raw(query, params...).Scan(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Gagal mengambil data activity",
			"details": err.Error(),
		})
		return
	}

	// ============================
	// Buat dokumen_link dinamis
	// ============================
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + c.Request.Host

	for i := range activities {
		if activities[i].DokumenLink != nil {
			fullURL := baseURL + *activities[i].DokumenLink
			activities[i].DokumenLink = &fullURL
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  activities,
	})
}
