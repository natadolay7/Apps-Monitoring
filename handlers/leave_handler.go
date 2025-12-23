package handlers

import (
	"net/http"
	"time"

	"api_patroliku_docker/database"
	"api_patroliku_docker/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LeaveHandler struct {
	DB *gorm.DB
}

func NewLeaveHandler() *LeaveHandler {
	return &LeaveHandler{
		DB: database.GetDB(),
	}
}
func (h *LeaveHandler) SaveLeave(c *gin.Context) {
	var req models.LeaveSaveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	var ubc UserBranchCompany

	sql := `
		SELECT 
			c.id AS company_id,
			b.id AS branch_id
		FROM user_tad_information uti
		LEFT JOIN branch b ON b.id = uti.branch_id
		LEFT JOIN company c ON c.id = b.company_id
		WHERE uti.user_id = ?
	`

	if err := h.DB.Raw(sql, req.UserTadID).Scan(&ubc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data company & branch",
			"error":   err.Error(),
		})
		return
	}

	if ubc.CompanyID == 0 || ubc.BranchID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Company atau branch tidak ditemukan untuk user ini",
		})
		return
	}

	// ===== Parse tanggal (opsional) =====
	var dateStart time.Time
	var dateEnd time.Time

	if req.DateStart != "" {
		t, err := time.Parse("2006-01-02", req.DateStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Format date_start harus YYYY-MM-DD",
			})
			return
		}
		dateStart = t
	}

	if req.DateEnd != "" {
		t, err := time.Parse("2006-01-02", req.DateEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Format date_end harus YYYY-MM-DD",
			})
			return
		}
		dateEnd = t
	}

	// ===== Documents (opsional) =====
	var documents models.JSONMap
	if req.Document != "" {
		documents = make(models.JSONMap)
		documents["document"] = req.Document
	}

	leave := models.Leave{
		UserTadID:          req.UserTadID,
		LeaveTypeID:        req.LeaveTypeID,
		CompanyID:          ubc.CompanyID, // ✅ AUTO
		BranchID:           ubc.BranchID,  // ✅ AUTO
		UserClientID:       req.UserClientID,
		UserClientBranchID: req.UserClientBranchID,
		UserCoordinatorID:  req.UserCoordinatorID,
		TypeLeave:          req.TypeLeave,
		Code:               req.Code,
		Documents:          documents,
		DateRequest:        time.Now(),
		DateStart:          dateStart,
		DateEnd:            dateEnd,
		Note:               req.Note,
		Status:             "1", // 1 = Pending
	}

	if err := h.DB.Create(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menyimpan leave",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Leave berhasil disimpan",
		"data": gin.H{
			"id":           leave.ID,
			"user_tad_id":  leave.UserTadID,
			"type_leave":   leave.TypeLeave,
			"date_start":   req.DateStart,
			"date_end":     req.DateEnd,
			"status":       "Pending",
			"date_request": leave.DateRequest,
		},
	})
}

type UserBranchCompany struct {
	CompanyID uint `gorm:"column:company_id"`
	BranchID  uint `gorm:"column:branch_id"`
}
