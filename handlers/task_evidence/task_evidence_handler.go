package handlers

import (
	"encoding/json"
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

type TaskEvidenceHandler struct {
	DB *gorm.DB
}

func NewTaskEvidenceHandler() *TaskEvidenceHandler {
	return &TaskEvidenceHandler{
		DB: database.GetDB(),
	}
}

// UploadTaskEvidence - Upload before dan after photo dalam satu function
func (h *TaskEvidenceHandler) UploadTaskEvidence(c *gin.Context) {
	// ===== Parse form data =====
	taskAssignIDStr := c.PostForm("task_assign_id")
	userTadIDStr := c.PostForm("user_tad_id")
	note := c.PostForm("note")
	taskCondition := c.PostForm("task_condition") // "1" = before, "2" = after
	status := c.PostForm("status")                // Status task evidence

	if taskAssignIDStr == "" || userTadIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "task_assign_id dan user_tad_id wajib diisi",
		})
		return
	}

	taskAssignID, err := strconv.Atoi(taskAssignIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "task_assign_id tidak valid",
		})
		return
	}

	userTadID, err := strconv.Atoi(userTadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": true,
		})
		return
	}

	// ===== Get branch_id dari user =====
	var branchID int
	err = h.DB.Raw(`
		SELECT branch_id 
		FROM user_tad_information 
		WHERE user_id = ?
	`, userTadID).Scan(&branchID).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "branch user tidak ditemukan",
		})
		return
	}

	// ===== Handle file upload (bisa null) =====
	var photoURLs []string
	var evidenceType string // "before" atau "after"

	beforeFile, beforeErr := c.FormFile("before_photo")
	afterFile, afterErr := c.FormFile("after_photo")

	if beforeErr == nil && beforeFile != nil {
		// Upload before photo
		evidenceType = "before"
		photoURL, err := h.saveUploadedFile(c, beforeFile, "before")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Gagal menyimpan before photo",
			})
			return
		}
		photoURLs = append(photoURLs, photoURL)
	}

	if afterErr == nil && afterFile != nil {
		// Upload after photo
		evidenceType = "after"
		photoURL, err := h.saveUploadedFile(c, afterFile, "after")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Gagal menyimpan after photo",
			})
			return
		}
		photoURLs = append(photoURLs, photoURL)
	}

	// ===== Jika tidak ada file yang diupload =====
	if len(photoURLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Minimal satu foto harus diupload (before_photo atau after_photo)",
		})
		return
	}

	// ===== Cek apakah sudah ada task_evidence untuk task_assign_id ini =====
	type ExistingEvidence struct {
		ID           int
		BeforePhotos string
		AfterPhotos  string
	}

	var existing ExistingEvidence

	h.DB.Raw(`
		SELECT id, before_photos, after_photos 
		FROM task_evidence 
		WHERE task_assign_id = ?
		LIMIT 1
	`, taskAssignID).Scan(&existing)

	// ===== Convert photo URLs to JSON =====
	photosJSON, _ := json.Marshal(photoURLs)

	// ===== Insert atau Update ke database =====
	if existing.ID == 0 {
		// INSERT baru
		query := `
			INSERT INTO task_evidence 
			(task_assign_id, user_tad_id, branch_id, task_condition, before_photos, after_photos, status, note, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		var beforePhotosJSON, afterPhotosJSON interface{}
		if evidenceType == "before" {
			beforePhotosJSON = photosJSON
			afterPhotosJSON = nil
		} else {
			beforePhotosJSON = nil
			afterPhotosJSON = photosJSON
		}

		if err := h.DB.Exec(
			query,
			taskAssignID,
			userTadID,
			branchID,
			taskCondition,
			beforePhotosJSON,
			afterPhotosJSON,
			status,
			note,
			time.Now(),
			time.Now(),
		).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":         true,
				"message":       "Gagal menyimpan task evidence",
				"error_details": err.Error(),
			})
			return
		}
	} else {
		// UPDATE existing
		query := `
			UPDATE task_evidence 
			SET 
		`

		params := []interface{}{}

		if evidenceType == "before" {
			// Update before photos
			var finalBeforePhotos []string

			// Parse existing before photos
			if existing.BeforePhotos != "" {
				json.Unmarshal([]byte(existing.BeforePhotos), &finalBeforePhotos)
			}

			// Append new photos
			finalBeforePhotos = append(finalBeforePhotos, photoURLs...)

			// Convert to JSON
			finalBeforeJSON, _ := json.Marshal(finalBeforePhotos)

			query += " before_photos = ?, "
			params = append(params, finalBeforeJSON)

		} else if evidenceType == "after" {
			// Update after photos
			var finalAfterPhotos []string

			// Parse existing after photos
			if existing.AfterPhotos != "" {
				json.Unmarshal([]byte(existing.AfterPhotos), &finalAfterPhotos)
			}

			// Append new photos
			finalAfterPhotos = append(finalAfterPhotos, photoURLs...)

			// Convert to JSON
			finalAfterJSON, _ := json.Marshal(finalAfterPhotos)

			query += " after_photos = ?, "
			params = append(params, finalAfterJSON)

			// Jika upload after photo, update status task_assign menjadi completed
			h.updateTaskAssignStatus(taskAssignID, "completed")
		}

		query += " note = ?, updated_at = ? WHERE id = ?"
		params = append(params, note, time.Now(), existing.ID)

		if err := h.DB.Exec(query, params...).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":         true,
				"message":       "Gagal update task evidence",
				"error_details": err.Error(),
			})
			return
		}
	}

	// ===== Response =====
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Task evidence berhasil disimpan",
		"data": gin.H{
			"evidence_type":  evidenceType,
			"photo_urls":     photoURLs,
			"task_assign_id": taskAssignID,
		},
	})
}

// UpdateTaskAssign - Update status dan data task_assign
func (h *TaskEvidenceHandler) UpdateTaskAssign(c *gin.Context) {
	taskAssignIDStr := c.Param("id")
	if taskAssignIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID task assign wajib diisi",
		})
		return
	}

	taskAssignID, err := strconv.Atoi(taskAssignIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID task assign tidak valid",
		})
		return
	}

	// ===== Parse request body =====
	var updateData struct {
		Status    string  `json:"status"`
		Note      *string `json:"note"`
		StartTime *string `json:"start_time"`
		EndTime   *string `json:"end_time"`
		NoteTad   *string `json:"note_tad"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request data",
		})
		return
	}

	// ===== Build update query dynamically =====
	query := "UPDATE task_assign SET updated_at = ?"
	params := []interface{}{time.Now()}

	if updateData.Status != "" {
		query += ", status = ?"
		params = append(params, updateData.Status)
	}

	if updateData.Note != nil {
		query += ", note = ?"
		params = append(params, *updateData.Note)
	}

	if updateData.StartTime != nil {
		query += ", start_time = ?"
		params = append(params, *updateData.StartTime)
	}

	if updateData.EndTime != nil {
		query += ", end_time = ?"
		params = append(params, *updateData.EndTime)
	}

	if updateData.NoteTad != nil {
		query += ", note = ?" // note_tad disimpan di field note
		params = append(params, *updateData.NoteTad)
	}

	query += " WHERE id = ?"
	params = append(params, taskAssignID)

	// ===== Execute update =====
	result := h.DB.Exec(query, params...)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"message":       "Gagal update task assign",
			"error_details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Task assign tidak ditemukan",
		})
		return
	}

	// ===== Response =====
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Task assign berhasil diupdate",
		"data": gin.H{
			"task_assign_id": taskAssignID,
			"status":         updateData.Status,
			"updated_at":     time.Now(),
		},
	})
}

// Helper function to save uploaded file
func (h *TaskEvidenceHandler) saveUploadedFile(c *gin.Context, fileHeader *multipart.FileHeader, evidenceType string) (string, error) {
	// Create upload directory
	uploadDir := "uploads/task_evidence/" + evidenceType
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	// Generate unique filename
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(fileHeader.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
		return "", err
	}

	// Generate full URL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	photoURL := scheme + "://" + c.Request.Host + "/" + filePath
	return photoURL, nil
}

// Helper function to update task_assign status
func (h *TaskEvidenceHandler) updateTaskAssignStatus(taskAssignID int, status string) error {
	query := "UPDATE task_assign SET status = ?, updated_at = ? WHERE id = ?"
	return h.DB.Exec(query, status, time.Now(), taskAssignID).Error
}

// Helper function to get task evidence by task_assign_id
func (h *TaskEvidenceHandler) GetTaskEvidence(c *gin.Context) {
	taskAssignIDStr := c.Param("id")
	if taskAssignIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID task assign wajib diisi",
		})
		return
	}

	taskAssignID, err := strconv.Atoi(taskAssignIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID task assign tidak valid",
		})
		return
	}

	type TaskEvidenceResponse struct {
		ID            int      `json:"id"`
		TaskAssignID  int      `json:"task_assign_id"`
		BeforePhotos  []string `json:"before_photos"`
		AfterPhotos   []string `json:"after_photos"`
		Status        string   `json:"status"`
		Note          string   `json:"note"`
		TaskCondition string   `json:"task_condition"`
		CreatedAt     string   `json:"created_at"`
		UpdatedAt     string   `json:"updated_at"`
	}

	var evidence TaskEvidenceResponse

	query := `
		SELECT 
			id,
			task_assign_id,
			before_photos,
			after_photos,
			status,
			note,
			task_condition,
			created_at,
			updated_at
		FROM task_evidence
		WHERE task_assign_id = ?
		LIMIT 1
	`

	if err := h.DB.Raw(query, taskAssignID).Scan(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Gagal mengambil task evidence",
		})
		return
	}

	if evidence.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Task evidence tidak ditemukan",
		})
		return
	}

	// Parse JSON strings to arrays

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Task evidence berhasil diambil",
		"data":    evidence,
	})
}
