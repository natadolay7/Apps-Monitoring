// models/attendance_response.go
package models

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
