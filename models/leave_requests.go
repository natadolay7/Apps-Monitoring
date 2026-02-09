package models

import "time"

type LeaveRequest struct {
	ID          int       `gorm:"primaryKey"`
	UserID      int       `gorm:"column:user_id"`
	BranchID    int       `gorm:"column:branch_id"`
	LeaveTypeID int       `gorm:"column:leave_type_id"`
	StartDate   time.Time `gorm:"column:start_date"`
	EndDate     time.Time `gorm:"column:end_date"`
	TotalDays   int       `gorm:"column:total_days"`
	Reason      string    `gorm:"column:reason"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (LeaveRequest) TableName() string {
	return "leave_requests"
}
