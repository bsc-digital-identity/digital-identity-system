package outbox

import (
	"api/src/database"
	"api/src/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxRetries = 5

type OutboxRepository interface {
	GetEvent(uuid.UUID) (model.OutboxEvent, error)
	NewEvent(identityId, schemaId uuid.UUID, reqestMsg string) (uuid.UUID, error)
	GetUnprocessedEvents() ([]model.OutboxEvent, error)
	MarkEventAsProcessed(uuid.UUID) error
	UpdateRetryValue(eventId uuid.UUID) error
	WithTx(*gorm.DB) OutboxRepository
}

type outboxRepository struct {
	db *gorm.DB
}

// WithTx allows the repository to run operations within a transaction
func (or *outboxRepository) WithTx(tx *gorm.DB) OutboxRepository {
	if tx == nil {
		return or
	}
	repo := &outboxRepository{
		db: tx,
	}
	return repo
}

func NewRepo() OutboxRepository {
	return &outboxRepository{db: database.GetDatabaseConnection()}
}

func (or *outboxRepository) GetEvent(eventId uuid.UUID) (model.OutboxEvent, error) {
	var event model.OutboxEvent
	result := or.db.Unscoped().First(&event, "event_id = ? AND processed_at IS NULL", eventId.String())
	return event, result.Error
}

func (or *outboxRepository) NewEvent(identityId, schemaId uuid.UUID, reqestMsg string) (uuid.UUID, error) {
	eventId, err := uuid.NewRandom()
	if err != nil {
		return eventId, err
	}

	event := &model.OutboxEvent{
		EventId:        eventId.String(),
		IdentityId:     identityId.String(),
		SchemaId:       schemaId.String(),
		RequestMessage: reqestMsg,
		ToProcess:      false,
		Retry:          0,
	}

	result := or.db.Create(event)
	return eventId, result.Error
}

func (or *outboxRepository) GetUnprocessedEvents() ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	result := or.db.Model(&model.OutboxEvent{}).Where("to_process = ?", true).Find(&events)
	return events, result.Error
}

func (or *outboxRepository) MarkEventAsProcessed(eventId uuid.UUID) error {
	return or.db.Transaction(func(tx *gorm.DB) error {
		// Use WithTx to ensure we're using the transaction
		repo := or.WithTx(tx)

		// Get the event first to ensure it exists
		_, err := repo.GetEvent(eventId)
		if err != nil {
			return err
		}

		// Set processed_at to current time (soft delete)
		result := tx.Model(&model.OutboxEvent{}).
			Where("event_id = ?", eventId.String()).
			Update("processed_at", time.Now())
		return result.Error
	})
}

// TODO: optimize this query
func (or *outboxRepository) UpdateRetryValue(eventId uuid.UUID) error {
	return or.db.Transaction(func(tx *gorm.DB) error {
		repo := or.WithTx(tx)

		// Get the event first
		res, err := repo.GetEvent(eventId)
		if err != nil {
			return err
		}

		// Update retry count
		err = tx.Model(&model.OutboxEvent{}).
			Where("event_id = ?", eventId.String()).
			Update("retry", res.Retry+1).
			Error
		if err != nil {
			return err
		}

		// If max retries reached, mark as processed
		if res.Retry >= maxRetries {
			return repo.MarkEventAsProcessed(eventId)
		}

		return nil
	})
}
