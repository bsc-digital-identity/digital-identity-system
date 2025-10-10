package integration

import (
	"api/src/model"
	"api/src/outbox"
	"api/test/integration/utils"
	"fmt"
	"os"
	"testing"
	"time"

	"pkg-common/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	testDB *gorm.DB
)

func setupTestDB(t *testing.T) *gorm.DB {
	if testDB == nil {
		db := utils.GetTestDB(t)
		if err := db.AutoMigrate(&model.OutboxEvent{}); err != nil {
			t.Fatalf("Failed to migrate test database: %v", err)
		}
		testDB = db
	}
	// Start transaction for test
	tx := testDB.Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})
	return tx
}

func TestMain(m *testing.M) {
	// Setup logger
	config := logger.GlobalLoggerConfig{
		Args: []logger.LoggerArg{
			{Key: "service", Value: "outbox-test"},
		},
	}
	logger.InitDefaultLogger(config)

	// Setup
	t := &testing.T{}
	db := utils.SetupTestDB(t)
	if db == nil {
		os.Exit(1)
	}
	testDB = db

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testDB.Migrator().DropTable(&model.OutboxEvent{}); err != nil {
		fmt.Printf("Failed to cleanup test database: %v\n", err)
	}

	utils.CleanupTestDB(t)
	os.Exit(code)
}

func createTestEvent(t *testing.T, db *gorm.DB, toProcess bool) uuid.UUID {
	eventId, err := uuid.NewRandom()
	assert.NoError(t, err)

	identityId, err := uuid.NewRandom()
	assert.NoError(t, err)

	schemaId, err := uuid.NewRandom()
	assert.NoError(t, err)

	event := model.OutboxEvent{
		EventId:        eventId.String(),
		IdentityId:     identityId.String(),
		SchemaId:       schemaId.String(),
		Retry:          0,
		ToProcess:      toProcess,
		RequestMessage: "test message",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result := db.Create(&event)
	assert.NoError(t, result.Error)

	return eventId
}

func TestGetEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := outbox.NewRepoWithDB(db)

	// Test getting non-existent event
	nonExistentId, err := uuid.NewRandom()
	assert.NoError(t, err)
	_, err = repo.GetEvent(nonExistentId)
	assert.Error(t, err)

	// Test getting existing event
	eventId := createTestEvent(t, db, false)
	event, err := repo.GetEvent(eventId)
	assert.NoError(t, err)
	assert.Equal(t, eventId.String(), event.EventId)
}

func TestNewEvent(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		identityId, err := uuid.NewRandom()
		assert.NoError(t, err)

		schemaId, err := uuid.NewRandom()
		assert.NoError(t, err)

		eventId, err := repo.NewEvent(identityId, schemaId, "test message")
		assert.NoError(t, err)

		// Verify event was created with correct values
		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, identityId.String(), event.IdentityId)
		assert.Equal(t, schemaId.String(), event.SchemaId)
		assert.Equal(t, "test message", event.RequestMessage)
		assert.Equal(t, 0, event.Retry)
		assert.False(t, event.ToProcess)
		assert.NotEmpty(t, event.CreatedAt)
		assert.NotEmpty(t, event.UpdatedAt)
		assert.Empty(t, event.ProcessedAt)
	})

	t.Run("creation with empty message", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		identityId, _ := uuid.NewRandom()
		schemaId, _ := uuid.NewRandom()

		eventId, err := repo.NewEvent(identityId, schemaId, "")
		assert.NoError(t, err)

		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Empty(t, event.RequestMessage)
	})

	t.Run("duplicate event ID", func(t *testing.T) {
		db := setupTestDB(t)

		identityId, _ := uuid.NewRandom()
		schemaId, _ := uuid.NewRandom()
		existingEventId := createTestEvent(t, db, false)

		// Try to create event with same ID
		event := model.OutboxEvent{
			EventId:        existingEventId.String(),
			IdentityId:     identityId.String(),
			SchemaId:       schemaId.String(),
			RequestMessage: "test message",
		}

		err := db.Create(&event).Error
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestGetUnprocessedEvents(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		events, err := repo.GetUnprocessedEvents()
		assert.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("only processed events", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		_ = createTestEvent(t, db, false)
		_ = createTestEvent(t, db, false)

		events, err := repo.GetUnprocessedEvents()
		assert.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("single unprocessed event", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		eventId := createTestEvent(t, db, true)
		_ = createTestEvent(t, db, false)

		events, err := repo.GetUnprocessedEvents()
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, eventId.String(), events[0].EventId)
		assert.True(t, events[0].ToProcess)
	})

	t.Run("multiple mixed events", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		// Create mix of processed and unprocessed events
		unprocessedIds := make([]string, 0)
		for i := 0; i < 3; i++ {
			if i%2 == 0 {
				eventId := createTestEvent(t, db, true)
				unprocessedIds = append(unprocessedIds, eventId.String())
			} else {
				_ = createTestEvent(t, db, false)
			}
		}

		events, err := repo.GetUnprocessedEvents()
		assert.NoError(t, err)
		assert.Len(t, events, 2)

		// Verify all returned events are unprocessed and match our created IDs
		eventIds := make([]string, 0)
		for _, event := range events {
			assert.True(t, event.ToProcess)
			eventIds = append(eventIds, event.EventId)
		}
		assert.ElementsMatch(t, unprocessedIds, eventIds)
	})
}

func TestMarkEventAsProcessed(t *testing.T) {
	db := setupTestDB(t)
	repo := outbox.NewRepoWithDB(db)

	eventId := createTestEvent(t, db, true)

	// Mark as processed
	err := repo.MarkEventAsProcessed(eventId)
	assert.NoError(t, err)

	// Verify event was deleted
	_, err = repo.GetEvent(eventId)
	assert.Error(t, err)
}

func TestUpdateRetryValue(t *testing.T) {
	t.Run("successful retry increments", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		eventId := createTestEvent(t, db, true)

		// Test updating retry count
		for i := 1; i <= outbox.MaxRetries+1; i++ {
			err := repo.UpdateRetryValue(eventId)
			assert.NoError(t, err)

			if i <= outbox.MaxRetries {
				event, err := repo.GetEvent(eventId)
				assert.NoError(t, err)
				assert.Equal(t, i, event.Retry)
			} else {
				// After max retries, event should be marked as processed (deleted)
				_, err := repo.GetEvent(eventId)
				assert.Error(t, err)
			}
		}
	})

	t.Run("retry on non-existent event", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		nonExistentId, _ := uuid.NewRandom()
		err := repo.UpdateRetryValue(nonExistentId)
		assert.Error(t, err)
	})

	t.Run("concurrent retry updates", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		eventId := createTestEvent(t, db, true)

		// Simulate concurrent updates
		repo1 := outbox.NewRepoWithDB(db)
		repo2 := outbox.NewRepoWithDB(db)

		err1 := repo1.UpdateRetryValue(eventId)
		assert.NoError(t, err1)

		err2 := repo2.UpdateRetryValue(eventId)
		assert.NoError(t, err2)

		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, 2, event.Retry)
	})

	t.Run("retry on already processed event", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		eventId := createTestEvent(t, db, true)
		err := repo.MarkEventAsProcessed(eventId)
		assert.NoError(t, err)

		err = repo.UpdateRetryValue(eventId)
		assert.Error(t, err)
	})
}

func TestConcurrentOperations(t *testing.T) {
	t.Run("concurrent mark as processed", func(t *testing.T) {
		db := setupTestDB(t)
		eventId := createTestEvent(t, db, true)

		// Create two repository instances to simulate concurrent operations
		repo1 := outbox.NewRepoWithDB(db)
		repo2 := outbox.NewRepoWithDB(db)

		// First mark as processed should succeed
		err1 := repo1.MarkEventAsProcessed(eventId)
		assert.NoError(t, err1)

		// Second attempt should fail as event is already processed
		err2 := repo2.MarkEventAsProcessed(eventId)
		assert.Error(t, err2)
	})
}

func TestTransactionRollback(t *testing.T) {
	t.Run("rollback on error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		eventId := createTestEvent(t, db, true)

		// Begin a transaction in a closure to ensure proper cleanup
		err := db.Transaction(func(tx *gorm.DB) error {
			txRepo := repo.WithTx(tx)

			// Update event within transaction
			if err := txRepo.UpdateRetryValue(eventId); err != nil {
				return err
			}

			// Verify retry count is updated in transaction
			event, err := txRepo.GetEvent(eventId)
			if err != nil {
				return err
			}
			assert.Equal(t, 1, event.Retry)

			// Force rollback by returning error
			return fmt.Errorf("forced rollback")
		})

		// Expect an error since we forced a rollback
		assert.Error(t, err)
		assert.Equal(t, "forced rollback", err.Error())

		// Verify retry count is unchanged in main DB
		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, 0, event.Retry)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("very large request message", func(t *testing.T) {
		db := setupTestDB(t)

		identityId, _ := uuid.NewRandom()
		schemaId, _ := uuid.NewRandom()

		// Create a large message (1MB)
		largeMessage := make([]byte, 1024*1024)
		for i := range largeMessage {
			largeMessage[i] = byte('a')
		}

		repo := outbox.NewRepoWithDB(db)
		eventId, err := repo.NewEvent(identityId, schemaId, string(largeMessage))
		assert.NoError(t, err)

		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, string(largeMessage), event.RequestMessage)
	})

	t.Run("zero uuid values", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		zeroUUID := uuid.Nil
		eventId, err := repo.NewEvent(zeroUUID, zeroUUID, "test message")
		assert.NoError(t, err)

		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, zeroUUID.String(), event.IdentityId)
		assert.Equal(t, zeroUUID.String(), event.SchemaId)
	})

	t.Run("unicode message content", func(t *testing.T) {
		db := setupTestDB(t)
		repo := outbox.NewRepoWithDB(db)

		identityId, _ := uuid.NewRandom()
		schemaId, _ := uuid.NewRandom()
		message := "Test 测试 テスト 테스트 اختبار"

		eventId, err := repo.NewEvent(identityId, schemaId, message)
		assert.NoError(t, err)

		event, err := repo.GetEvent(eventId)
		assert.NoError(t, err)
		assert.Equal(t, message, event.RequestMessage)
	})
}
