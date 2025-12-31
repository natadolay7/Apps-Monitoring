package handlers

import (
	"math"
	"net/http"
	"strconv"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB *gorm.DB
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		DB: database.GetDB(),
	}
}

func (h *TaskHandler) GetTaskByUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id wajib diisi",
		})
		return
	}

	// ===== user_id dari param =====
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id tidak valid",
		})
		return
	}

	// ===== pagination =====
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// ===== total data =====
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM task_assign ta 
		LEFT JOIN task t ON ta.task_id = t.id 
		LEFT JOIN task_type tt ON tt.id = t.task_type_id 
		WHERE ta.user_tad_id = ? AND t.deleted_at IS NULL
	`

	if err := h.DB.Raw(countQuery, userID).Scan(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghitung data",
			"error":   err.Error(),
		})
		return
	}

	// ===== data list =====

	// Buat struct untuk response baru
	type TaskAssignResponse struct {
		ID        int     `json:"id"`
		Status    string  `json:"status"`
		TitleTask string  `json:"title_task"`
		Type      string  `json:"type"`
		Note      *string `json:"note"`
		StartTime *string `json:"start_time"`
		EndTime   *string `json:"end_time"`
	}

	var taskAssigns []TaskAssignResponse

	dataQuery := `
		SELECT 
			ta.id,
			ta.status,
			t.name AS title_task,
			tt.name AS type,
			ta.note,
			ta.start_time,
			ta.end_time
		FROM task_assign ta 
		LEFT JOIN task t ON ta.task_id = t.id 
		LEFT JOIN task_type tt ON tt.id = t.task_type_id 
		WHERE ta.user_tad_id = ? AND t.deleted_at IS NULL
		ORDER BY ta.id DESC
		LIMIT ? OFFSET ?
	`

	if err := h.DB.Raw(dataQuery, userID, limit, offset).
		Scan(&taskAssigns).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil task assign",
			"error":   err.Error(),
		})
		return
	}

	totalPage := int(math.Ceil(float64(total) / float64(limit)))

	// ===== response =====
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Task assign berhasil diambil",
		"data":    taskAssigns,
		"metadata": gin.H{
			"user_id":    userID,
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
			"has_data":   len(taskAssigns) > 0,
		},
	})
}

func (h *TaskHandler) GetTask(c *gin.Context) {

	// ===== pagination =====
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// ===== total data =====
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM task_assign ta 
		LEFT JOIN task t ON ta.task_id = t.id 
		LEFT JOIN task_type tt ON tt.id = t.task_type_id 
		WHERE t.deleted_at IS NULL
	`

	if err := h.DB.Raw(countQuery).
		Scan(&total).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menghitung data",
			"error":   err.Error(),
		})
		return
	}

	// ===== data list =====
	type TaskAssignResponse struct {
		ID        int     `json:"id"`
		Status    string  `json:"status"`
		TitleTask string  `json:"title_task"`
		Type      string  `json:"type"`
		Note      *string `json:"note"`
		StartTime *string `json:"start_time"`
		EndTime   *string `json:"end_time"`
	}

	var taskAssigns []TaskAssignResponse

	dataQuery := `
		SELECT 
			ta.id,
			ta.status,
			t.name AS title_task,
			tt.name AS type,
			ta.note,
			ta.start_time,
			ta.end_time
		FROM task_assign ta 
		LEFT JOIN task t ON ta.task_id = t.id 
		LEFT JOIN task_type tt ON tt.id = t.task_type_id 
		WHERE t.deleted_at IS NULL
		ORDER BY ta.id DESC
		LIMIT ? OFFSET ?
	`

	if err := h.DB.Raw(dataQuery, limit, offset).
		Scan(&taskAssigns).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil task assign",
			"error":   err.Error(),
		})
		return
	}

	totalPage := int(math.Ceil(float64(total) / float64(limit)))

	// ===== response =====
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Task assign berhasil diambil",
		"data":    taskAssigns,
		"metadata": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
			"has_data":   len(taskAssigns) > 0,
		},
	})
}

func (h *TaskHandler) GetTaskDetail(c *gin.Context) {
	// ===== Get task_assign_id dari path parameter =====
	taskAssignIDStr := c.Param("id")
	if taskAssignIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID task wajib diisi",
		})
		return
	}

	taskAssignID, err := strconv.Atoi(taskAssignIDStr)
	if err != nil || taskAssignID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "ID task tidak valid",
		})
		return
	}

	// ===== Struct untuk response detail task =====
	type TaskDetailResponse struct {
		ID        int     `json:"id"`
		Status    string  `json:"status"`
		TitleTask string  `json:"title_task"`
		Type      string  `json:"type"`
		Note      *string `json:"note"`
		StartTime *string `json:"start_time"`
		EndTime   *string `json:"end_time"`
	}

	var taskDetail TaskDetailResponse

	// ===== Query untuk mengambil detail task =====
	detailQuery := `
		SELECT 
			ta.id,
			ta.status,
			t.name AS title_task,
			tt.name AS type,
			ta.note,
			ta.start_time,
			ta.end_time,
			ta.end_time, 
			te.before_photos, 
			te.note  as note_pengerjaan
		FROM task_assign ta 
		LEFT JOIN task t ON ta.task_id = t.id 
		LEFT JOIN task_type tt ON tt.id = t.task_type_id 
		LEFT JOIN task_evidence te ON te.task_assign_id  = ta.id
		WHERE ta.id = ? AND t.deleted_at IS NULL
		LIMIT 1
	`

	if err := h.DB.Raw(detailQuery, taskAssignID).
		Scan(&taskDetail).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil detail task",
			"error":   err.Error(),
		})
		return
	}

	// ===== Cek apakah data ditemukan =====
	if taskDetail.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Task tidak ditemukan",
		})
		return
	}

	// ===== Response =====
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Detail task berhasil diambil",
		"data":    taskDetail,
	})
}
