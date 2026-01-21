package handlers

import (
	"net/http"
	"time"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	DB *gorm.DB
}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		DB: database.GetDB(),
	}
}

// ============================
// ADD / UPDATE PROFILE BY user_id
// ============================
func (h *ProfileHandler) StoreOrUpdateProfile(c *gin.Context) {

	// Ambil user_id dari JWT
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User tidak terautentikasi",
		})
		return
	}
	userID := userIDRaw.(int)

	// Ambil form / json data
	namaLengkap := c.PostForm("nama_lengkap")
	tempatLahir := c.PostForm("tempat_lahir")
	tanggalLahir := c.PostForm("tanggal_lahir") // YYYY-MM-DD
	jenisKelamin := c.PostForm("jenis_kelamin")
	alamat := c.PostForm("alamat")
	noHP := c.PostForm("no_hp")
	email := c.PostForm("email")

	if namaLengkap == "" || tanggalLahir == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "nama_lengkap dan tanggal_lahir wajib diisi",
		})
		return
	}

	// Cek apakah profile sudah ada
	var count int64
	h.DB.Raw(`SELECT COUNT(*) FROM biodata WHERE user_id = ?`, userID).Scan(&count)

	now := time.Now()

	if count == 0 {
		// ============================
		// INSERT
		// ============================
		query := `
			INSERT INTO biodata 
			(user_id, nama_lengkap, tempat_lahir, tanggal_lahir, jenis_kelamin, alamat, no_hp, email, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		if err := h.DB.Exec(
			query,
			userID,
			namaLengkap,
			tempatLahir,
			tanggalLahir,
			jenisKelamin,
			alamat,
			noHP,
			email,
			now,
			now,
		).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Gagal menambah profile",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"message": "Profile berhasil ditambahkan",
		})

	} else {
		// ============================
		// UPDATE
		// ============================
		query := `
			UPDATE biodata SET
				nama_lengkap = ?,
				tempat_lahir = ?,
				tanggal_lahir = ?,
				jenis_kelamin = ?,
				alamat = ?,
				no_hp = ?,
				email = ?,
				updated_at = ?
			WHERE user_id = ?
		`

		if err := h.DB.Exec(
			query,
			namaLengkap,
			tempatLahir,
			tanggalLahir,
			jenisKelamin,
			alamat,
			noHP,
			email,
			now,
			userID,
		).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Gagal update profile",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"message": "Profile berhasil diperbarui",
		})
	}
}
