package handlers

import (
	"api_patroliku_docker/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnnouncementsHandler struct {
	DB *gorm.DB
}

func NewAnnouncementsHandler() *AnnouncementsHandler {
	return &AnnouncementsHandler{
		DB: database.GetDB(),
	}
}

func (h *AnnouncementsHandler) GetActiveAnnouncementsByBranch(c *gin.Context) {
	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "branch_id wajib diisi",
		})
		return
	}

	type AnnouncementResponse struct {
		ID        int       `json:"id"`
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		StartDate string    `json:"start_date"`
		EndDate   string    `json:"end_date"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}

	var data []AnnouncementResponse

	query := `
		SELECT 
			a.id,
			a.title,
			a.content,
			a.start_date,
			a.end_date,
			a.status,
			a.created_at
		FROM announcements a
		WHERE a.branch_id = ?
		AND a.status = 'active'
		AND CURRENT_TIMESTAMP BETWEEN a.start_date AND a.end_date
		ORDER BY a.created_at DESC
	`

	if err := h.DB.Raw(query, branchID).Scan(&data).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil announcements",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Announcements aktif berhasil diambil",
		"data":    data,
	})
}
