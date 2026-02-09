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

func (h *LeaveHandler) GetLeaveBudget(c *gin.Context) {

	// Ambil userID dari middleware
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User tidak ditemukan di token",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Format userID tidak valid",
		})
		return
	}

	year := time.Now().Year()

	var result LeaveBudgetResponse

	sql := `
		SELECT 
			u.id,
			u.name,
			u.email as username,
			uti.branch_id,
			uti.annual_leave_quota,
			COALESCE(SUM(lr.total_days),0) as total_used,
			(uti.annual_leave_quota - COALESCE(SUM(lr.total_days),0)) as remaining_leave
		FROM users u
		JOIN user_tad_information uti ON u.id = uti.user_id
		LEFT JOIN leave_requests lr 
			ON u.id = lr.user_id
			AND lr.status = 'approved'
			AND EXTRACT(YEAR FROM lr.start_date) = ?
		WHERE u.id = ?
		GROUP BY 
			u.id,
			u.name,
			u.email,
			uti.branch_id,
			uti.annual_leave_quota
	`

	if err := h.DB.Raw(sql, year, userID).Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil budget cuti",
			"error":   err.Error(),
		})
		return
	}

	if result.UserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Data cuti tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Berhasil mengambil budget cuti",
		"data":    result,
	})
}

type LeaveBudgetResponse struct {
	UserID           uint    `json:"user_id" gorm:"column:id"`
	Name             string  `json:"name"`
	Username         string  `json:"username" gorm:"column:username"`
	BranchID         uint    `json:"branch_id"`
	AnnualLeaveQuota int     `json:"annual_leave_quota"`
	TotalUsed        float64 `json:"total_used"`
	RemainingLeave   float64 `json:"remaining_leave"`
}

type LeaveCreateRequest struct {
	LeaveTypeID int    `json:"leave_type_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate     string `json:"end_date" binding:"required"`
	Reason      string `json:"reason"`
}

func (h *LeaveHandler) CreateLeave(c *gin.Context) {

	var req LeaveCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// ===== Ambil userID & branchID dari middleware =====
	userIDValue, _ := c.Get("userID")
	branchIDValue, _ := c.Get("branchID")

	userID := userIDValue.(int)
	branchID := branchIDValue.(int)

	// ===== Parse tanggal =====
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format start_date harus YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format end_date harus YYYY-MM-DD",
		})
		return
	}

	// ===== Validasi end_date tidak boleh sebelum start_date =====
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "end_date tidak boleh sebelum start_date",
		})
		return
	}

	// ===== Hitung total hari (inclusive) =====
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

	leave := models.LeaveRequest{
		UserID:      userID,
		BranchID:    branchID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   startDate,
		EndDate:     endDate,
		TotalDays:   totalDays,
		Reason:      req.Reason,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
		"message": "Leave berhasil diajukan",
		"data": gin.H{
			"id":         leave.ID,
			"start_date": leave.StartDate,
			"end_date":   leave.EndDate,
			"total_days": totalDays,
			"status":     leave.Status,
		},
	})
}
