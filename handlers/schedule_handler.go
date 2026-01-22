package handlers

import (
	"net/http"
	"time"

	"api_patroliku_docker/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScheduleHandler struct {
	DB *gorm.DB
}

func NewScheduleHandler() *ScheduleHandler {
	return &ScheduleHandler{
		DB: database.GetDB(),
	}
}

func (h *ScheduleHandler) GetJadwalByUser(c *gin.Context) {

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

	// ============================
	// Struct Raw Query
	// ============================
	type JadwalRaw struct {
		ID        int     `json:"id"`
		Shift     *string `json:"shift"`
		Day       int     `json:"day"`
		Holiday   int     `json:"holiday"`
		StartTime *string `json:"start_time"`
		EndTime   *string `json:"end_time"`
	}

	var rows []JadwalRaw

	// ============================
	// Query
	// ============================
	query := `
		SELECT 
			s.id,
			ss."name" AS shift,
			s."day",
			s.holiday,
			ss.start_time,
			ss.end_time
		FROM schedule s
		LEFT JOIN schedule_shift ss ON ss.id = s.schedule_shift_id
		WHERE s.users_id = ?
		ORDER BY s."day" ASC
	`

	if err := h.DB.Raw(query, userID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Gagal mengambil jadwal",
			"details": err.Error(),
		})
		return
	}

	// ============================
	// Mapping Hari
	// ============================
	hariMap := map[int]string{
		1: "Senin",
		2: "Selasa",
		3: "Rabu",
		4: "Kamis",
		5: "Jumat",
		6: "Sabtu",
		7: "Minggu",
	}

	// Tentukan awal minggu (Senin)
	now := time.Now()
	offset := int(now.Weekday())
	if offset == 0 {
		offset = 7
	}
	startOfWeek := now.AddDate(0, 0, -(offset - 1))

	// ============================
	// Response Final
	// ============================
	var result []gin.H

	for _, r := range rows {

		tanggal := startOfWeek.AddDate(0, 0, r.Day-1)

		status := "Kerja"
		if r.Holiday == 1 {
			status = "Libur"
		}

		result = append(result, gin.H{
			"id":         r.ID,
			"day_number": r.Day,
			"hari":       hariMap[r.Day],
			"tanggal":    tanggal.Format("2006-01-02"),
			"shift":      r.Shift,
			"start_time": r.StartTime,
			"end_time":   r.EndTime,
			"status":     status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"error": false,
		"data":  result,
	})
}
