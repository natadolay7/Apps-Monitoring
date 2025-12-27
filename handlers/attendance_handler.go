package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"api_patroliku_docker/database"
	"api_patroliku_docker/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AttendanceHandler struct {
	DB *gorm.DB
}

func NewAttendanceHandler() *AttendanceHandler {
	return &AttendanceHandler{
		DB: database.GetDB(),
	}
}

// GetAttendanceByUserAndDate - GET /api/v1/attendance
func (h *AttendanceHandler) GetAttendanceByUserAndDate(c *gin.Context) {
	userID := c.Query("user_id")
	date := c.Query("date")

	if userID == "" || date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user_id dan date wajib diisi",
		})
		return
	}

	var result models.AttendanceResponse
	var scheduleID sql.NullInt64

	query := `
		SELECT 
			ss.start_time as check_in,
			ss.end_time as check_out,
			u.name,
			s.date_check_in as date,
			b.latitude,
			b.longitude,
			b.radius,
			b.id as branch_id,
			s.id as id_schedule
		FROM schedule s 
		LEFT JOIN schedule_shift ss ON s.schedule_shift_id = ss.id
		LEFT JOIN users u ON s.users_id = u.id
		LEFT JOIN user_tad_information uti ON u.id = uti.user_id
		LEFT JOIN branch b ON b.id = uti.branch_id
		WHERE s.users_id = ?
		  AND s.date_check_in = ?
	`

	row := h.DB.Raw(query, userID, date).Row()

	err := row.Scan(
		&result.CheckIn,
		&result.CheckOut,
		&result.Name,
		&result.Date,
		&result.Latitude,
		&result.Longitude,
		&result.Radius,
		&result.BranchID,
		&scheduleID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "data tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "gagal mengambil data",
			"error":   err.Error(),
		})
		return
	}

	// Handle nullable schedule ID
	if scheduleID.Valid {
		result.ScheduleID = scheduleID.Int64
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    result,
	})
}

func (h *AttendanceHandler) SaveAttendance(c *gin.Context) {
	var req models.AttendanceSaveRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Validasi sederhana
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id harus diisi",
		})
		return
	}

	if req.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "date harus diisi (format: YYYY-MM-DD)",
		})
		return
	}

	if req.AttendanceStatus == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "attendance_status harus diisi (1=Hadir, 2=Sakit, 3=Izin)",
		})
		return
	}

	// Parse tanggal
	dateAttendance, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format tanggal harus YYYY-MM-DD",
			"example": "2025-12-13",
		})
		return
	}

	// Parse waktu check_in (jika ada)
	var checkInTime *time.Time
	if req.CheckIn != "" {
		checkIn := h.parseTimeString(req.CheckIn, dateAttendance)
		checkInTime = &checkIn
	}

	// Parse waktu check_out (jika ada)
	var checkOutTime *time.Time
	if req.CheckOut != "" {
		checkOut := h.parseTimeString(req.CheckOut, dateAttendance)
		checkOutTime = &checkOut
	}

	// Persiapan dokumen
	var documentsClock models.JSONMap = make(models.JSONMap)
	if req.DocumentsClockIn != "" {
		documentsClock["check_in_document"] = req.DocumentsClockIn
	}
	if req.DocumentsClockOut != "" {
		documentsClock["check_out_document"] = req.DocumentsClockOut
	}

	// Cek apakah sudah ada data untuk user dan tanggal ini
	var existingAttendance models.UserAttendance
	err = h.DB.Where("users_id = ? AND date_attendence = ?", req.UserID, dateAttendance).
		First(&existingAttendance).Error

	if err == gorm.ErrRecordNotFound {
		// CREATE: Data baru
		attendance := models.UserAttendance{
			UserID:            req.UserID,
			AttendanceStatus:  req.AttendanceStatus,
			CheckIn:           checkInTime,
			CheckOut:          checkOutTime,
			DateAttendance:    dateAttendance,
			LongitudeCheckIn:  req.LongitudeCheckIn,
			LatitudeCheckIn:   req.LatitudeCheckIn,
			LongitudeCheckOut: req.LongitudeCheckOut,
			LatitudeCheckOut:  req.LatitudeCheckOut,
			DocumentsClock:    documentsClock,
		}

		if err := h.DB.Create(&attendance).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Gagal menyimpan attendance",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "Attendance berhasil disimpan",
			"data":    h.mapAttendanceToResponse(attendance),
		})

	} else if err != nil {
		// Error saat cek data
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memeriksa data existing",
			"error":   err.Error(),
		})
		return

	} else {
		// UPDATE: Data sudah ada
		updateData := models.UserAttendance{
			AttendanceStatus:  req.AttendanceStatus,
			LongitudeCheckIn:  req.LongitudeCheckIn,
			LatitudeCheckIn:   req.LatitudeCheckIn,
			LongitudeCheckOut: req.LongitudeCheckOut,
			LatitudeCheckOut:  req.LatitudeCheckOut,
		}

		// Update check_in jika diberikan
		if checkInTime != nil {
			updateData.CheckIn = checkInTime
		}

		// Update check_out jika diberikan
		if checkOutTime != nil {
			updateData.CheckOut = checkOutTime
		}

		// Update documents jika ada
		if len(documentsClock) > 0 {
			if existingAttendance.DocumentsClock != nil {
				// Gabungkan dengan existing
				for k, v := range documentsClock {
					existingAttendance.DocumentsClock[k] = v
				}
				updateData.DocumentsClock = existingAttendance.DocumentsClock
			} else {
				updateData.DocumentsClock = documentsClock
			}
		}

		// Eksekusi update
		if err := h.DB.Model(&existingAttendance).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Gagal update attendance",
				"error":   err.Error(),
			})
			return
		}

		// Ambil data terbaru
		h.DB.First(&existingAttendance, existingAttendance.ID)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Attendance berhasil diupdate",
			"data":    h.mapAttendanceToResponse(existingAttendance),
			"metadata": gin.H{
				"is_update": true,
			},
		})
	}
}

// Helper untuk parse waktu string

// Helper untuk map ke response
func (h *AttendanceHandler) mapAttendanceToResponse(attendance models.UserAttendance) models.AttendanceSaveResponse {
	response := models.AttendanceSaveResponse{
		ID:               attendance.ID,
		UserID:           attendance.UserID,
		AttendanceStatus: attendance.AttendanceStatus,
		DateAttendance:   attendance.DateAttendance,
		LongitudeCheckIn: attendance.LongitudeCheckIn,
		LatitudeCheckIn:  attendance.LatitudeCheckIn,
		CreatedAt:        attendance.CreatedAt,
		UpdatedAt:        attendance.UpdatedAt,
	}

	// ScheduleID (opsional)
	if attendance.ScheduleID != nil && *attendance.ScheduleID > 0 {
		response.ScheduleID = attendance.ScheduleID
	}

	// CheckIn (opsional)
	if attendance.CheckIn != nil {
		response.CheckIn = *attendance.CheckIn
	}

	// CheckOut (opsional)
	if attendance.CheckOut != nil {
		response.CheckOut = *attendance.CheckOut
	}

	// Longitude/Latitude CheckOut (opsional)
	if attendance.LongitudeCheckOut != 0 {
		response.LongitudeCheckOut = attendance.LongitudeCheckOut
	}
	if attendance.LatitudeCheckOut != 0 {
		response.LatitudeCheckOut = attendance.LatitudeCheckOut
	}

	return response
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req models.CheckInRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Validasi
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id harus diisi",
		})
		return
	}

	if req.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "date harus diisi (format: YYYY-MM-DD)",
		})
		return
	}

	if req.CheckIn == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "check_in harus diisi (format: HH:MM atau HH:MM:SS)",
		})
		return
	}

	// Parse tanggal
	dateAttendance, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format tanggal harus YYYY-MM-DD",
			"example": "2025-12-13",
		})
		return
	}

	// Parse waktu check_in
	checkInTime := h.parseTimeString(req.CheckIn, dateAttendance)

	// Cek apakah sudah ada data untuk user dan tanggal ini
	var existingAttendance models.UserAttendance
	err = h.DB.Where("users_id = ? AND date_attendence = ?", req.UserID, dateAttendance).
		First(&existingAttendance).Error

	// Persiapan dokumen check-in
	var documentsClock models.JSONMap = make(models.JSONMap)
	if req.DocumentsClockIn != "" {
		documentsClock["check_in_document"] = req.DocumentsClockIn
	}

	tx := h.DB.Begin()

	if err == gorm.ErrRecordNotFound {
		// CREATE: Data baru (hanya check-in)
		attendance := models.UserAttendance{
			UserID:           req.UserID,
			AttendanceStatus: 1, // Default: Hadir
			CheckIn:          &checkInTime,
			DateAttendance:   dateAttendance,
			LongitudeCheckIn: req.LongitudeCheckIn,
			LatitudeCheckIn:  req.LatitudeCheckIn,
			DocumentsClock:   documentsClock,
		}

		if err := tx.Create(&attendance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Gagal menyimpan check-in",
				"error":   err.Error(),
			})
			return
		}

		tx.Commit()

		c.JSON(http.StatusCreated, gin.H{
			"status":  "success",
			"message": "Check-in berhasil disimpan",
			"data": gin.H{
				"id":                attendance.ID,
				"user_id":           attendance.UserID,
				"date":              attendance.DateAttendance.Format("2006-01-02"),
				"check_in":          attendance.CheckIn.Format("15:04:05"),
				"check_out":         nil,
				"attendance_status": "Hadir",
				"created_at":        attendance.CreatedAt,
			},
		})

	} else if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memeriksa data existing",
			"error":   err.Error(),
		})

	} else {
		// UPDATE: Data sudah ada, update check-in saja
		// Cek apakah sudah check-in hari ini
		if existingAttendance.CheckIn != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Anda sudah melakukan check-in hari ini",
				"data": gin.H{
					"check_in_time": existingAttendance.CheckIn.Format("15:04:05"),
				},
			})
			return
		}

		// Update check-in
		updates := map[string]interface{}{
			"check_in":           checkInTime,
			"longitude_check_in": req.LongitudeCheckIn,
			"latitude_check_in":  req.LatitudeCheckIn,
			"updated_at":         time.Now(),
		}

		// Update dokumen jika ada
		if len(documentsClock) > 0 {
			if existingAttendance.DocumentsClock != nil {
				existingAttendance.DocumentsClock["check_in_document"] = req.DocumentsClockIn
				updates["documents_clock__"] = existingAttendance.DocumentsClock
			} else {
				updates["documents_clock__"] = documentsClock
			}
		}

		if err := tx.Model(&existingAttendance).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Gagal update check-in",
				"error":   err.Error(),
			})
			return
		}

		tx.Commit()

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Check-in berhasil diupdate",
			"data": gin.H{
				"id":                existingAttendance.ID,
				"user_id":           existingAttendance.UserID,
				"date":              existingAttendance.DateAttendance.Format("2006-01-02"),
				"check_in":          checkInTime.Format("15:04:05"),
				"check_out":         existingAttendance.CheckOut,
				"attendance_status": "Hadir",
				"updated_at":        time.Now(),
			},
		})
	}
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	var req models.CheckOutRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Validasi
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id harus diisi",
		})
		return
	}

	if req.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "date harus diisi (format: YYYY-MM-DD)",
		})
		return
	}

	if req.CheckOut == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "check_out harus diisi (format: HH:MM atau HH:MM:SS)",
		})
		return
	}

	// Parse tanggal
	dateAttendance, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format tanggal harus YYYY-MM-DD",
			"example": "2025-12-13",
		})
		return
	}

	// Parse waktu check_out
	checkOutTime := h.parseTimeString(req.CheckOut, dateAttendance)

	// Cek apakah sudah ada data attendance untuk user dan tanggal ini
	var existingAttendance models.UserAttendance
	err = h.DB.Where("users_id = ? AND date_attendence = ?", req.UserID, dateAttendance).
		First(&existingAttendance).Error

	if err == gorm.ErrRecordNotFound {
		// Tidak ada data check-in, tidak bisa check-out
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Anda belum melakukan check-in hari ini",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal memeriksa data attendance",
			"error":   err.Error(),
		})
		return
	}

	// Cek apakah sudah check-out
	if existingAttendance.CheckOut != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Anda sudah melakukan check-out hari ini",
			"data": gin.H{
				"check_out_time": existingAttendance.CheckOut.Format("15:04:05"),
			},
		})
		return
	}

	// Persiapan dokumen check-out
	var documentsClock models.JSONMap
	if req.DocumentsClockOut != "" {
		documentsClock = make(models.JSONMap)
		documentsClock["check_out_document"] = req.DocumentsClockOut
	}

	tx := h.DB.Begin()

	// Update check-out
	updates := map[string]interface{}{
		"check_out":           checkOutTime,
		"longitude_check_out": req.LongitudeCheckOut,
		"latitude_check_out":  req.LatitudeCheckOut,
		"updated_at":          time.Now(),
	}

	// Update dokumen jika ada
	if len(documentsClock) > 0 {
		if existingAttendance.DocumentsClock != nil {
			existingAttendance.DocumentsClock["documents_clock_out"] = req.DocumentsClockOut
			updates["documents_clock_out"] = existingAttendance.DocumentsClock
		} else {
			updates["documents_clock_out"] = documentsClock
		}
	}

	// Hitung total jam kerja jika ada check-in
	var workDuration string
	if existingAttendance.CheckIn != nil {
		duration := checkOutTime.Sub(*existingAttendance.CheckIn)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		workDuration = fmt.Sprintf("%d jam %d menit", hours, minutes)
	}

	if err := tx.Model(&existingAttendance).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal menyimpan check-out",
			"error":   err.Error(),
		})
		return
	}

	tx.Commit()

	// Response
	response := gin.H{
		"id":                existingAttendance.ID,
		"user_id":           existingAttendance.UserID,
		"date":              existingAttendance.DateAttendance.Format("2006-01-02"),
		"check_in":          existingAttendance.CheckIn.Format("15:04:05"),
		"check_out":         checkOutTime.Format("15:04:05"),
		"attendance_status": "Hadir",
		"updated_at":        time.Now(),
	}

	// Tambahkan durasi kerja jika ada
	if workDuration != "" {
		response["work_duration"] = workDuration
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Check-out berhasil disimpan",
		"data":    response,
	})
}

// GetTodayAttendance - GET /api/v1/attendance/today
func (h *AttendanceHandler) GetTodayAttendance(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "user_id harus diisi",
		})
		return
	}

	var userIDUint uint
	fmt.Sscanf(userID, "%d", &userIDUint)

	today := time.Now().Format("2006-01-02")

	var attendance models.UserAttendance
	err := h.DB.Where("users_id = ? AND date_attendence = ?", userIDUint, today).
		First(&attendance).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Belum ada attendance hari ini",
			"data": gin.H{
				"has_check_in":  false,
				"has_check_out": false,
				"date":          today,
			},
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil data attendance",
			"error":   err.Error(),
		})
		return
	}

	response := gin.H{
		"id":                attendance.ID,
		"user_id":           attendance.UserID,
		"date":              attendance.DateAttendance.Format("2006-01-02"),
		"has_check_in":      attendance.CheckIn != nil,
		"has_check_out":     attendance.CheckOut != nil,
		"attendance_status": attendance.AttendanceStatus,
	}

	if attendance.CheckIn != nil {
		response["check_in"] = attendance.CheckIn.Format("15:04:05")
		response["check_in_location"] = gin.H{
			"latitude":  attendance.LatitudeCheckIn,
			"longitude": attendance.LongitudeCheckIn,
		}
	}

	if attendance.CheckOut != nil {
		response["check_out"] = attendance.CheckOut.Format("15:04:05")
		response["check_out_location"] = gin.H{
			"latitude":  attendance.LatitudeCheckOut,
			"longitude": attendance.LongitudeCheckOut,
		}

		// Hitung durasi kerja jika ada check-in dan check-out
		if attendance.CheckIn != nil {
			duration := attendance.CheckOut.Sub(*attendance.CheckIn)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			response["work_duration"] = fmt.Sprintf("%d jam %d menit", hours, minutes)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data attendance hari ini",
		"data":    response,
	})
}

// Helper untuk parse waktu string
func (h *AttendanceHandler) parseTimeString(timeStr string, date time.Time) time.Time {
	// Coba format HH:MM:SS
	var hour, min, sec int
	n, err := fmt.Sscanf(timeStr, "%d:%d:%d", &hour, &min, &sec)
	if err != nil || n < 2 {
		// Coba format HH:MM
		n, err = fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
		if err != nil || n < 2 {
			// Default ke waktu saat ini
			return time.Now()
		}
		sec = 0
	}

	// Buat waktu dengan tanggal dari parameter
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		hour,
		min,
		sec,
		0,
		time.Local,
	)
}
