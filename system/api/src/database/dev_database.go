package database

import (
	"api/src/model"
	"encoding/json"
	"pkg-common/logger"

	"github.com/google/uuid"
)

// TODO: should do something with this
func InitializeDatabaseForDev() error {
	db := GetDatabaseConnection()
	// Example: Insert admin if not exists
	admin_id, _ := uuid.NewRandom()
	admin := model.Identity{
		IdentityId:   admin_id.String(),
		IdentityName: "admin",
	}

	result := db.FirstOrCreate(&admin)
	if result.Error != nil {
		logger.Default().Error(result.Error, "Error inserting admin")
	}

	constraint := model.Constraint{
		Key:        "Age",
		Comparison: model.GreaterThan,
		Value:      18,
	}

	schema := model.Schema{
		Constraints: []model.Constraint{constraint},
	}

	serialized_schema, err := json.Marshal(schema)
	if err != nil {
		logger.Default().Error(err, "Error when serializing")
	}

	schema_id, _ := uuid.NewRandom()
	verifiable_schema := model.VerifiedSchema{
		SuperIdentityId: admin.Id,
		SchemaId:        schema_id.String(),
		Schema:          string(serialized_schema),
	}

	result = db.FirstOrCreate(&verifiable_schema)

	if result.Error != nil {
		logger.Default().Error(result.Error, "Error when inserting schema")
	}

	logger.Default().Infof("Created admin user with id: %s", admin_id.String())
	logger.Default().Infof("Created schema with Age constraint: %s", schema_id.String())
	return nil
}
