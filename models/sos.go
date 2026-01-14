package models

type SOSAlert struct {
	ID        uint    `gorm:"primaryKey;column:id"`
	SenderID  uint    `gorm:"column:sender_id"`
	BranchID  uint    `gorm:"column:branch_id"`
	Latitude  float64 `gorm:"column:latitude"`
	Longitude float64 `gorm:"column:longitude"`
	Message   string  `gorm:"column:message"`
	Status    string  `gorm:"column:status;default:sent"`
}

func (SOSAlert) TableName() string {
	return "sos_reports"
}

type SOSRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Message   string  `json:"message"`
}
