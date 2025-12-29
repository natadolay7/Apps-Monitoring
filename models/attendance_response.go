// models/attendance_response.go
package models

import "time"

type AttendanceResponse struct {
	CheckIn    string  `json:"check_in"`
	CheckOut   string  `json:"check_out"`
	Name       string  `json:"name"`
	Date       string  `json:"date"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Radius     int     `json:"radius"`
	BranchID   int64   `json:"branch_id"`
	ScheduleID int64   `json:"id_schedule"`
}

type AttendanceScheduleResponse struct {
	Shift        string    `json:"shift"`
	CheckinTime  time.Time `json:"checkin_time"`
	CheckoutTime time.Time `json:"checkout_time"`
	NamaBranch   string    `json:"nama_branch"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	UsersID      int       `json:"users_id"`
	Day          int       `json:"day"`
	Holiday      bool      `json:"holiday"`
	UserID       int       `json:"user_id"`
}

type BranchResponse struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ScheduleResponse struct {
	IDShift      int            `json:"id_shift"`
	IDSchedule   int            `json:"id_schedule"`
	Shift        string         `json:"shift"`
	CheckinTime  string         `json:"checkin_time"`
	CheckoutTime string         `json:"checkout_time"`
	Branch       BranchResponse `json:"branch"`
}

type AttendanceTodayResponse struct {
	Holiday  bool              `json:"holiday"`
	Schedule *ScheduleResponse `json:"schedule"`
}
