package logaudit

import (
	"api/src/model"
	"time"

	logger_message "pkg-common/utilities/logger"
)

const (
	apiServiceName = "api"
)

type LogAuditService interface {
	ProcessLogMessage(logMessage logger_message.LoggerMessage) error
	GetLogEntries(limit, offset int) ([]model.LogAuditEntry, error)
	GetLogEntriesByService(service string, limit, offset int) ([]model.LogAuditEntry, error)
	GetLogEntriesByLevel(level string, limit, offset int) ([]model.LogAuditEntry, error)
}

type logAuditService struct {
	repository LogAuditRepository
}

func NewLogAuditService(repository LogAuditRepository) LogAuditService {
	return &logAuditService{
		repository: repository,
	}
}

func (s *logAuditService) ProcessLogMessage(logMessage logger_message.LoggerMessage) error {
	logEntry := model.LogAuditEntry{
		Level:     logMessage.Level,
		Message:   logMessage.Message,
		Timestamp: time.Unix(logMessage.Timestamp.T, 0).UTC(),
		Service:   apiServiceName,
	}

	return s.repository.CreateLogEntry(logEntry)
}

func (s *logAuditService) GetLogEntries(limit, offset int) ([]model.LogAuditEntry, error) {
	return s.repository.GetLogEntries(limit, offset)
}

func (s *logAuditService) GetLogEntriesByService(service string, limit, offset int) ([]model.LogAuditEntry, error) {
	return s.repository.GetLogEntriesByService(service, limit, offset)
}

func (s *logAuditService) GetLogEntriesByLevel(level string, limit, offset int) ([]model.LogAuditEntry, error) {
	return s.repository.GetLogEntriesByLevel(level, limit, offset)
}
