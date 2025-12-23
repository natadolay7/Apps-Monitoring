package models

import "time"

type Leave struct {
	ID                 uint      `gorm:"primaryKey;column:id"`
	LeaveTypeID        uint      `gorm:"column:leave_type_id"`
	CompanyID          uint      `gorm:"column:company_id"`
	BranchID           uint      `gorm:"column:branch_id"`
	UserTadID          uint      `gorm:"column:user_tad_id"`
	UserClientID       uint      `gorm:"column:user_client_id"`
	UserClientBranchID uint      `gorm:"column:user_client_branch_id"`
	UserCoordinatorID  uint      `gorm:"column:user_coordinator_id"`
	TypeLeave          string    `gorm:"column:type_leave"`
	Code               string    `gorm:"column:code"`
	Documents          JSONMap   `gorm:"column:documents;type:json"`
	DateRequest        time.Time `gorm:"column:date_request"`
	DateStart          time.Time `gorm:"column:date_start"`
	DateEnd            time.Time `gorm:"column:date_end"`
	Note               string    `gorm:"column:note"`
	NoteApproval       string    `gorm:"column:note_approval"`
	Status             string    `gorm:"column:status"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time `gorm:"index"`
}

func (Leave) TableName() string {
	return "leave_new"
}
