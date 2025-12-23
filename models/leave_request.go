package models

type LeaveSaveRequest struct {
	UserTadID uint `json:"user_tad_id" binding:"required"`

	LeaveTypeID       uint `json:"leave_type_id"`
	CompanyID         uint `json:"company_id"`
	BranchID          uint `json:"branch_id"`
	UserClientID      uint `json:"user_client_id"`
	UserCoordinatorID uint `json:"user_coordinator_id"`

	TypeLeave string `json:"type_leave"`
	Code      string `json:"code"`

	DateStart          string `json:"date_start"` // optional
	DateEnd            string `json:"date_end"`   // optional
	Note               string `json:"note"`
	Document           string `json:"document"`
	UserClientBranchID uint   `gorm:"column:user_client_branch"`
}
