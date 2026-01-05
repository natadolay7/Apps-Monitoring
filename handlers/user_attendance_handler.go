package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserAttendanceHandler struct {
	DB *gorm.DB
}

func NewUserAttendanceHandler() *UserAttendanceHandler {
	return &UserAttendanceHandler{
		DB: database.GetDB(),
	}
}

// =====================================================
// GET USER ATTENDANCE HARI INI
// =====================================================
func (h *UserAttendanceHandler) GetUserAttendanceToday(c *gin.Context) {
	// ===== ambil user_id =====
	userIDStr := c.Query("user_id")
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

	// ===== tanggal hari ini (handler) =====
	today := time.Now().Format("2006-01-02")

	// ===== struct response =====
	type UserAttendanceResponse struct {
		HasAttendance bool `json:"has_attendance"`

		UserName string `json:"user_name"`

		CheckInUser  *time.Time `json:"check_in_user"`
		CheckOutUser *time.Time `json:"check_out_user"`

		CheckInTime  string `json:"check_in_time"`
		CheckOutTime string `json:"check_out_time"`

		BranchID         *int    `json:"branch_id"`
		NamaShift        *string `json:"nama_shift"`
		CheckInSchedule  *string `json:"check_in_schedule"`
		CheckOutSchedule *string `json:"check_out_schedule"`
		CreatedDate      string  `json:"created_date"`

		IsLate       bool   `json:"is_late"`
		LateSeconds  int64  `json:"late_seconds"`
		LateMinutes  int    `json:"late_minutes"`
		LateDuration string `json:"late_duration"`
		LateHuman    string `json:"late_human_readable"`
	}

	var data UserAttendanceResponse

	// ===== query =====
	query := `
		SELECT 
			u.name AS user_name,
			ua.check_in AS check_in_user, 
			ua.check_out AS check_out_user,
			b.id AS branch_id,
			ss.name AS nama_shift,
			ss.start_time AS check_in_schedule,
			ss.end_time AS check_out_schedule,
			DATE(ua.created_at) AS created_date
		FROM user_attendence ua 
		LEFT JOIN schedule s ON s.id = ua.schedule_id
		LEFT JOIN schedule_shift ss ON ss.id = s.schedule_shift_id 
		LEFT JOIN users u ON u.id = ua.users_id
		LEFT JOIN user_tad_information uti ON u.id = uti.user_id 
		LEFT JOIN branch b ON b.id = uti.branch_id 
		WHERE u.id = ?
		  AND DATE(ua.created_at) = ?
		LIMIT 1
	`

	if err := h.DB.Raw(query, userID, today).
		Scan(&data).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "gagal mengambil user attendance",
			"error":   err.Error(),
		})
		return
	}

	// =====================================================
	// RESPONSE SAAT DATA KOSONG (BELUM ABSEN)
	// =====================================================
	if data.UserName == "" {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Attendance hari ini belum tersedia",
			"data": gin.H{
				"has_attendance": false,

				"user_name":      "",
				"check_in_user":  nil,
				"check_out_user": nil,

				"check_in_time":  "",
				"check_out_time": "",

				"branch_id":          nil,
				"nama_shift":         nil,
				"check_in_schedule":  nil,
				"check_out_schedule": nil,
				"created_date":       "",

				"is_late":             false,
				"late_seconds":        0,
				"late_minutes":        0,
				"late_duration":       "00:00:00",
				"late_human_readable": "Belum absen",
			},
		})
		return
	}

	// =====================================================
	// DATA ADA
	// =====================================================
	data.HasAttendance = true

	// ===== convert jam check-in & check-out =====
	if data.CheckInUser != nil {
		data.CheckInTime = data.CheckInUser.Format("15:04:05")
	}

	if data.CheckOutUser != nil {
		data.CheckOutTime = data.CheckOutUser.Format("15:04:05")
	}

	// ===== default nilai =====
	data.LateDuration = "00:00:00"
	data.LateHuman = "Tepat waktu"

	// ===== hitung keterlambatan =====
	if data.CheckInUser != nil && data.CheckInSchedule != nil {
		scheduleTime, err := time.Parse(
			"2006-01-02 15:04:05",
			today+" "+*data.CheckInSchedule,
		)
		if err == nil && data.CheckInUser.After(scheduleTime) {
			diff := data.CheckInUser.Sub(scheduleTime)

			data.IsLate = true
			data.LateSeconds = int64(diff.Seconds())
			data.LateMinutes = int(diff.Minutes())
			data.LateDuration = formatDurationToHMS(diff)
			data.LateHuman = formatDurationHuman(diff)
		}
	}

	// ===== response sukses =====
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Attendance hari ini berhasil diambil",
		"data":    data,
	})
}

// =====================================================
// HELPER FUNCTIONS
// =====================================================
func formatDurationToHMS(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func formatDurationHuman(d time.Duration) string {
	if d <= 0 {
		return "Tepat waktu"
	}

	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	parts := []string{}

	if hours > 0 {
		parts = append(parts, strconv.Itoa(hours)+" jam")
	}
	if minutes > 0 {
		parts = append(parts, strconv.Itoa(minutes)+" menit")
	}
	if seconds > 0 {
		parts = append(parts, strconv.Itoa(seconds)+" detik")
	}

	return strings.Join(parts, " ")
}
