package handlers

import (
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MasterPatroliHandler struct {
	DB *gorm.DB
}

func NewMasterPatroliHandler() *MasterPatroliHandler {
	return &MasterPatroliHandler{
		DB: database.GetDB(),
	}
}

func (h *MasterPatroliHandler) GetMasterPatroliByID(c *gin.Context) {
	// ===== ambil id dari path =====
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "id wajib diisi",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "id tidak valid",
		})
		return
	}

	// ===== struct response =====
	type MasterPatroliResponse struct {
		ID         int    `json:"id"`
		Kode       string `json:"kode"`
		NamaLokasi string `json:"nama_lokasi"`
	}

	var data MasterPatroliResponse

	// ===== query =====
	query := `
		SELECT 
			id,
			kode,
			nama_lokasi
		FROM master_patroli
		WHERE id = ?
		LIMIT 1
	`

	if err := h.DB.Raw(query, id).
		Scan(&data).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil master patroli",
			"error":   err.Error(),
		})
		return
	}

	// ===== cek data ditemukan =====
	if data.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Master patroli tidak ditemukan",
		})
		return
	}

	// ===== response =====
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Master patroli berhasil diambil",
		"data":    data,
	})
}

func (h *MasterPatroliHandler) StorePatroliReport(c *gin.Context) {
	// ===== ambil form data =====
	userIDStr := c.PostForm("user_id")
	idPatroliStr := c.PostForm("id_patroli")
	deskripsi := c.PostForm("deskripsi")
	latitude := c.PostForm("latitude")
	longitude := c.PostForm("longitude")

	if userIDStr == "" || idPatroliStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id dan id_patroli wajib diisi",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id tidak valid",
		})
		return
	}

	idPatroli, err := strconv.Atoi(idPatroliStr)
	if err != nil || idPatroli <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "id_patroli tidak valid",
		})
		return
	}

	// ===== upload image =====
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "image wajib diupload",
		})
		return
	}

	imageURL, err := h.savePatroliImage(c, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "gagal menyimpan image",
		})
		return
	}

	// ===== insert database =====
	query := `
		INSERT INTO patroli_report
		(user_id, id_patroli, deskripsi, image_url, created_at, updated_at , latitude , longitude)
		VALUES (?, ?, ?, ?, ?, ? , ? , ?)
	`

	if err := h.DB.Exec(
		query,
		userID,
		idPatroli,
		deskripsi,
		imageURL,
		time.Now(),
		time.Now(),
		latitude,
		longitude,
	).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "gagal menyimpan patroli report",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Patroli report berhasil disimpan",
		"data": gin.H{
			"user_id":    userID,
			"id_patroli": idPatroli,
			"image_url":  imageURL,
		},
	})
}

func (h *MasterPatroliHandler) savePatroliImage(
	c *gin.Context,
	fileHeader *multipart.FileHeader,
) (string, error) {

	uploadDir := "uploads/patroli_report"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	filename := strconv.FormatInt(time.Now().UnixNano(), 10) +
		filepath.Ext(fileHeader.Filename)

	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
		return "", err
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + c.Request.Host + "/" + filePath, nil
}

func (h *MasterPatroliHandler) ListPatroliReport(c *gin.Context) {
	// ===== ambil query param =====
	userIDStr := c.Query("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id wajib diisi",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id tidak valid",
		})
		return
	}

	// ===== default tanggal =====
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // 1 bulan ke belakang

	// ===== jika filter tanggal dikirim =====
	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "format start_date harus YYYY-MM-DD",
			})
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "format end_date harus YYYY-MM-DD",
			})
			return
		}

		// supaya end_date sampai akhir hari
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	// ===== struct response =====
	type PatroliReportResponse struct {
		ID         int       `json:"id"`
		Deskripsi  string    `json:"deskripsi"`
		ImageURL   string    `json:"image_url"`
		CreatedAt  time.Time `json:"created_at"`
		NamaLokasi string    `json:"nama_lokasi"`
		Latitude   string    `json:"latitude"`
		Longitude  string    `json:"longitude"`
	}

	var data []PatroliReportResponse

	// ===== query =====
	query := `
		SELECT 
			pr.id,
			pr.deskripsi,
			pr.image_url,
			pr.created_at,
			mp.nama_lokasi,
			pr.latitude,
			pr.longitude
		FROM patroli_report pr
		LEFT JOIN master_patroli mp ON pr.id_patroli = mp.id
		WHERE pr.user_id = ?
		  AND pr.created_at BETWEEN ? AND ?
		ORDER BY pr.created_at DESC
	`

	if err := h.DB.Raw(
		query,
		userID,
		startDate,
		endDate,
	).Scan(&data).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "gagal mengambil data patroli",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "List patroli berhasil diambil",
		"filter": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"data": data,
	})
}
