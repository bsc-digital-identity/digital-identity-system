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
}

type outboxRepository struct {
	db *gorm.DB
}

func NewRepo() OutboxRepository {
	return &outboxRepository{db: database.GetDatabaseConnection()}
}

func (or *outboxRepository) GetEvent(eventId uuid.UUID) (model.OutboxEvent, error) {
	var event model.OutboxEvent
	result := or.db.First(&event, "event_id = ?", eventId.String())
	return event, result.Error
}

func (or *outboxRepository) NewEvent(identityId, schemaId uuid.UUID, reqestMsg string) (uuid.UUID, error) {
	eventId, err := uuid.NewRandom()
	if err != nil {
		return eventId, err
	}

	result := or.db.Create(model.OutboxEvent{
		EventId:        eventId.String(),
		IdentityId:     identityId.String(),
		SchemaId:       schemaId.String(),
		Retry:          0,
		ToProcess:      false,
		RequestMessage: reqestMsg,
		CreatedAt:      time.Now().String(),
	})

	return eventId, result.Error
}

func (or *outboxRepository) GetUnprocessedEvents() ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	result := or.db.Select(&events).Where("should_process = 1")
	return events, result.Error
}

func (or *outboxRepository) MarkEventAsProcessed(eventId uuid.UUID) error {
	return or.db.Delete(&model.OutboxEvent{}).Where("event_id = ?", eventId.String()).Error
}

// TODO: optimize this query
func (or *outboxRepository) UpdateRetryValue(eventId uuid.UUID) error {
	res, err := or.GetEvent(eventId)
	if err != nil {
		return err
	}

	if res.Retry < maxRetries {
		return or.db.
			Where(&model.OutboxEvent{}, "event_id = ?", eventId.String()).
			Update("retry", res.Retry+1).Error
	}

	// should look into these events manually
	err = or.db.
		Where(&model.OutboxEvent{}, "event_id = ?", eventId.String()).
		Update("retry", res.Retry+1).Error
	if err != nil {
		return err
	}

	return or.MarkEventAsProcessed(eventId)
}
