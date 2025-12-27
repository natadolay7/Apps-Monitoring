package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// NullTime untuk handle nullable time
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the Scanner interface
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return nil
	case string:
		// Parse berbagai format waktu
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"15:04:05",
			"15:04",
			time.RFC3339,
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				nt.Time, nt.Valid = t, true
				return nil
			}
		}

		// Jika gagal parse, coba parse sebagai waktu saja
		if strings.Contains(v, ":") {
			parts := strings.Split(v, ":")
			if len(parts) >= 2 {
				hour := 0
				minute := 0
				fmt.Sscanf(parts[0], "%d", &hour)
				fmt.Sscanf(parts[1], "%d", &minute)
				t := time.Date(0, 1, 1, hour, minute, 0, 0, time.UTC)
				nt.Time, nt.Valid = t, true
				return nil
			}
		}

		nt.Valid = false
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into NullTime", value)
	}
}

// Value implements the driver Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON marshals NullTime to JSON
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(nt.Time.Format("15:04"))
}

// UnmarshalJSON unmarshals NullTime from JSON
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil {
		nt.Valid = false
		return nil
	}

	// Parse time from string
	t, err := time.Parse("15:04", *s)
	if err != nil {
		// Try other formats
		t, err = time.Parse("15:04:05", *s)
		if err != nil {
			return err
		}
	}
	nt.Time = t
	nt.Valid = true
	return nil
}

// TimeString untuk waktu dalam format string (HH:MM)
type TimeString string

// Scan implements the Scanner interface for TimeString
func (ts *TimeString) Scan(value interface{}) error {
	if value == nil {
		*ts = ""
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*ts = TimeString(v.Format("15:04"))
		return nil
	case string:
		// Parse dan format ulang ke HH:MM
		t, err := time.Parse("15:04:05", v)
		if err != nil {
			// Coba format lain
			t, err = time.Parse("15:04", v)
			if err != nil {
				*ts = TimeString(v) // Return as is
				return nil
			}
		}
		*ts = TimeString(t.Format("15:04"))
		return nil
	case []byte:
		return ts.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into TimeString", value)
	}
}

// AttendanceResponse model untuk response absen
// type AttendanceResponse struct {
// 	CheckIn    TimeString `json:"check_in"`
// 	CheckOut   TimeString `json:"check_out"`
// 	Name       string     `json:"name"`
// 	Date       string     `json:"date"` // Format: YYYY-MM-DD
// 	Latitude   float64    `json:"latitude"`
// 	Longitude  float64    `json:"longitude"`
// 	Radius     float64    `json:"radius"`
// 	BranchID   uint       `json:"branch_id"`
// 	UserID     uint       `json:"user_id"`
// 	IdSchedule uint       `json:"id_schedule"`
// }

// AttendanceRequest model untuk parameter request
type AttendanceRequest struct {
	UserID uint   `form:"user_id" binding:"required"`
	Date   string `form:"date" binding:"required"` // Format: YYYY-MM-DD
}

// AttendanceSummary untuk summary absen
type AttendanceSummary struct {
	Date          string  `json:"date"`
	UserName      string  `json:"user_name"`
	CheckIn       *string `json:"check_in"`    // Format: HH:MM
	CheckOut      *string `json:"check_out"`   // Format: HH:MM
	ShiftStart    *string `json:"shift_start"` // Format: HH:MM
	ShiftEnd      *string `json:"shift_end"`   // Format: HH:MM
	WorkHours     string  `json:"work_hours"`
	LateMinutes   int     `json:"late_minutes"`
	EarlyOut      int     `json:"early_out_minutes"`
	Status        string  `json:"status"` // on_time, late, early_out, no_check_in, no_check_out
	BranchName    string  `json:"branch_name"`
	BranchAddress string  `json:"branch_address"`
}

// AttendanceHistory untuk history absen
type AttendanceHistory struct {
	Date              string  `json:"date"`
	Name              string  `json:"name"`
	ScheduledCheckIn  string  `json:"scheduled_check_in"`
	ScheduledCheckOut string  `json:"scheduled_check_out"`
	ActualCheckIn     *string `json:"actual_check_in"`
	ActualCheckOut    *string `json:"actual_check_out"`
	BranchName        string  `json:"branch_name"`
	AttendanceStatus  string  `json:"attendance_status"`
}

// AttendanceRange untuk data range
type AttendanceRange struct {
	Date      string  `json:"date"`
	CheckIn   *string `json:"check_in"`
	CheckOut  *string `json:"check_out"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status"`
}

// AttendanceSaveRequest - Request untuk menyimpan absen
type AttendanceSaveRequest struct {
	UserID            uint    `json:"user_id" binding:"required"`
	Date              string  `json:"date" binding:"required"`
	AttendanceStatus  int     `json:"attendance_status" binding:"required"`
	CheckIn           string  `json:"check_in,omitempty"`
	CheckOut          string  `json:"check_out,omitempty"`
	LongitudeCheckIn  float64 `json:"longitude_check_in,omitempty"`
	LatitudeCheckIn   float64 `json:"latitude_check_in,omitempty"`
	LongitudeCheckOut float64 `json:"longitude_check_out,omitempty"`
	LatitudeCheckOut  float64 `json:"latitude_check_out,omitempty"`
	DocumentsClockIn  string  `json:"documents_clock_in,omitempty"`
	DocumentsClockOut string  `json:"documents_clock_out,omitempty"`
}

// AttendanceSaveResponse - Response setelah menyimpan absen
type AttendanceSaveResponse struct {
	ID                uint      `json:"id"`
	ScheduleID        *uint     `json:"schedule_id"`
	UserID            uint      `json:"user_id"`
	AttendanceStatus  int       `json:"attendance_status"`
	CheckIn           time.Time `json:"check_in,omitempty"`
	CheckOut          time.Time `json:"check_out,omitempty"`
	DateAttendance    time.Time `json:"date_attendance"`
	LongitudeCheckIn  float64   `json:"longitude_check_in,omitempty"`
	LatitudeCheckIn   float64   `json:"latitude_check_in,omitempty"`
	LongitudeCheckOut float64   `json:"longitude_check_out,omitempty"`
	LatitudeCheckOut  float64   `json:"latitude_check_out,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserAttendance - Model untuk tabel user_attendence
type UserAttendance struct {
	ID                uint       `gorm:"column:id;primaryKey;autoIncrement"`
	ScheduleID        *uint      `gorm:"column:schedule_id"`
	UserID            uint       `gorm:"column:users_id"`
	AttendanceStatus  int        `gorm:"column:attendence_status_id"`
	CheckIn           *time.Time `gorm:"column:check_in"`
	CheckOut          *time.Time `gorm:"column:check_out"`
	DateAttendance    time.Time  `gorm:"column:date_attendence"`
	LongitudeCheckIn  float64    `gorm:"column:longitude_check_in"`
	LatitudeCheckIn   float64    `gorm:"column:latitude_check_in"`
	LongitudeCheckOut float64    `gorm:"column:longitude_check_out"`
	LatitudeCheckOut  float64    `gorm:"column:latitude_check_out"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt         *time.Time `gorm:"column:deleted_at"`
	DocumentsClock    JSONMap    `gorm:"column:documents_clock_out;type:json"`
}

// JSONMap - Custom type untuk field JSON
type JSONMap map[string]interface{}

// Scan - Implement scanner untuk JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// Value - Implement valuer untuk JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (UserAttendance) TableName() string {
	return "user_attendence"
}

// Di file models/attendance.go

// CheckInRequest - Request untuk check-in
type CheckInRequest struct {
	UserID           uint    `json:"user_id" binding:"required"`
	Date             string  `json:"date" binding:"required"` // Format: YYYY-MM-DD
	CheckIn          string  `json:"check_in" binding:"required"`
	LongitudeCheckIn float64 `json:"longitude_check_in" binding:"required"`
	LatitudeCheckIn  float64 `json:"latitude_check_in" binding:"required"`
	DocumentsClockIn string  `json:"documents_clock_in,omitempty"` // Foto/selfie saat check-in
}

// CheckOutRequest - Request untuk check-out
type CheckOutRequest struct {
	UserID            uint    `json:"user_id" binding:"required"`
	Date              string  `json:"date" binding:"required"` // Format: YYYY-MM-DD
	CheckOut          string  `json:"check_out" binding:"required"`
	LongitudeCheckOut float64 `json:"longitude_check_out" binding:"required"`
	LatitudeCheckOut  float64 `json:"latitude_check_out" binding:"required"`
	DocumentsClockOut string  `json:"documents_clock_out,omitempty"` // Foto/selfie saat check-out
}
