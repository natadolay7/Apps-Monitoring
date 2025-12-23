package handlers

// import (
// 	"api_patroliku/database"
// 	"encoding/json"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// type TaskEvidenceHandler struct {
// 	DB *gorm.DB
// }

// func NewTaskEvidenceHandler() *TaskEvidenceHandler {
// 	return &TaskEvidenceHandler{
// 		DB: database.GetDB(),
// 	}
// }

// func (h *TaskEvidenceHandler) UploadBeforePhoto(c *gin.Context) {
// 	taskListID, _ := strconv.Atoi(c.PostForm("task_list_id"))
// 	userTadID, _ := strconv.Atoi(c.PostForm("user_tad_id"))
// 	note := c.PostForm("note")

// 	var branchID int
// 	err := h.DB.Raw(`
// 		SELECT branch_id
// 		FROM user_tad_information
// 		WHERE user_id = ?
// 	`, userTadID).Scan(&branchID).Error

// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"message": "branch user tidak ditemukan",
// 			"error":   err.Error(),
// 		})
// 		return
// 	}

// 	file, err := c.FormFile("photo")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "foto wajib diupload"})
// 		return
// 	}

// 	// ===== simpan file =====
// 	// simpan file
// 	uploadDir := "uploads/task_evidence/before"
// 	_ = os.MkdirAll(uploadDir, os.ModePerm)

// 	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(file.Filename)
// 	filePath := filepath.Join(uploadDir, filename)

// 	if err := c.SaveUploadedFile(file, filePath); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal simpan foto"})
// 		return
// 	}

// 	// 🔥 BUAT URL FULL DOMAIN / IP
// 	scheme := "http"
// 	if c.Request.TLS != nil {
// 		scheme = "https"
// 	}

// 	photoURL := scheme + "://" + c.Request.Host + "/" + filePath

// 	beforePhotos, _ := json.Marshal([]string{photoURL})

// 	// ===== insert DB =====
// 	query := `
// 		INSERT INTO task_evidence
// 		(task_list_id, task_condition , user_tad_id, branch_id , before_photos, status, note, created_at)
// 		VALUES (?, ?, ? , ? , ?, ?, ?, ?)
// 	`

// 	if err := h.DB.Exec(
// 		query,
// 		taskListID,
// 		"2",
// 		userTadID,
// 		branchID,
// 		beforePhotos,
// 		"1",
// 		note,
// 		time.Now(),
// 	).Error; err != nil {

// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "gagal simpan evidence",
// 			"error":   err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":  "success",
// 		"message": "before photo berhasil disimpan",
// 		"photo":   photoURL,
// 	})
// }

// func (h *TaskEvidenceHandler) UploadAfterPhoto(c *gin.Context) {
// 	taskListID, _ := strconv.Atoi(c.PostForm("task_list_id"))
// 	userTadID, _ := strconv.Atoi(c.PostForm("user_tad_id"))

// 	file, err := c.FormFile("photo")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "foto wajib diupload"})
// 		return
// 	}

// 	// ===== simpan file =====
// 	uploadDir := "uploads/task_evidence/after"
// 	_ = os.MkdirAll(uploadDir, os.ModePerm)

// 	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(file.Filename)
// 	filePath := filepath.Join(uploadDir, filename)

// 	if err := c.SaveUploadedFile(file, filePath); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "gagal simpan foto"})
// 		return
// 	}

// 	// ===== buat URL full =====
// 	scheme := "http"
// 	if c.Request.TLS != nil {
// 		scheme = "https"
// 	}

// 	photoURL := scheme + "://" + c.Request.Host + "/" + filePath
// 	afterPhotos, _ := json.Marshal([]string{photoURL})

// 	// ===== tanggal hari ini =====
// 	today := time.Now().Format("2006-01-02")

// 	// ===== update DB =====
// 	query := `
// 		UPDATE task_evidence
// 		SET after_photos = ?, status = ?, updated_at = ?
// 		WHERE task_list_id = ?
// 		AND user_tad_id = ?
// 		AND DATE(created_at) = ?
// 	`

// 	if err := h.DB.Exec(
// 		query,
// 		afterPhotos,
// 		"2",
// 		time.Now(),
// 		taskListID,
// 		userTadID,
// 		today,
// 	).Error; err != nil {

// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"message": "gagal update after photo",
// 			"error":   err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":  "success",
// 		"message": "after photo berhasil diupload",
// 		"photo":   photoURL,
// 	})
// }
