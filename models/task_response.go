package models

type TaskResponse struct {
	ID         uint   `json:"id" gorm:"column:id"`
	TaskName   string `json:"task_name" gorm:"column:task_name"`
	TaskType   string `json:"task_type" gorm:"column:task_type"`
	TaskTypeId int    `json:"task_type_id" gorm:"column:task_type_id"`
	StartTime  string `json:"start_time" gorm:"column:start_time"` // ✅ FIX
	EndTime    string `json:"end_time" gorm:"column:end_time"`     // ✅ FIX
	Note       string `json:"note" gorm:"column:note"`
}
