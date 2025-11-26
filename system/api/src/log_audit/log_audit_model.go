package logaudit

import (
	"time"

	"gorm.io/gorm"
)

type LogAuditEntry struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Level     string         `gorm:"type:varchar(10);not null;index" json:"level"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Timestamp time.Time      `gorm:"type:timestamp;not null;index" json:"timestamp"`
	Service   string         `gorm:"type:varchar(50);not null;index" json:"service"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LogAuditEntry) TableName() string {
	return "log_audit_entries"
}
