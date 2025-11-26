package logaudit

import (
	"api/src/database"

	"gorm.io/gorm"
)

type LogAuditRepository interface {
	CreateLogEntry(entry LogAuditEntry) error
	GetLogEntries(limit, offset int) ([]LogAuditEntry, error)
	GetLogEntriesByService(service string, limit, offset int) ([]LogAuditEntry, error)
	GetLogEntriesByLevel(level string, limit, offset int) ([]LogAuditEntry, error)
}

type logAuditRepository struct {
	db *gorm.DB
}

func NewLogAuditRepository() LogAuditRepository {
	return &logAuditRepository{
		db: database.GetDatabaseConnection(),
	}
}

func (r *logAuditRepository) CreateLogEntry(entry LogAuditEntry) error {
	result := r.db.Create(&entry)
	return result.Error
}

func (r *logAuditRepository) GetLogEntries(limit, offset int) ([]LogAuditEntry, error) {
	var entries []LogAuditEntry
	result := r.db.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&entries)
	return entries, result.Error
}

func (r *logAuditRepository) GetLogEntriesByService(service string, limit, offset int) ([]LogAuditEntry, error) {
	var entries []LogAuditEntry
	result := r.db.Where("service = ?", service).Order("timestamp DESC").Limit(limit).Offset(offset).Find(&entries)
	return entries, result.Error
}

func (r *logAuditRepository) GetLogEntriesByLevel(level string, limit, offset int) ([]LogAuditEntry, error) {
	var entries []LogAuditEntry
	result := r.db.Where("level = ?", level).Order("timestamp DESC").Limit(limit).Offset(offset).Find(&entries)
	return entries, result.Error
}
