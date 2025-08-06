package main

import (
	"api/src/model"
	"encoding/json"
	"pkg-common/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiConfigJson struct {
	loggerConf logger.LoggerConfigJson `json:"logger"`
}

type ApiConfig struct {
	LoggerConf logger.LoggerConfig
}

func (acj ApiConfigJson) ConvertToDomain() ApiConfig {
	return ApiConfig{
		LoggerConf: acj.loggerConf.ConvertToDomain(),
	}
}

func InitializeDev(db *gorm.DB) error {
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

	constraint := model.Constraint[int]{
		Key:        "Age",
		Comparison: model.GreaterThan,
		Value:      18,
	}

	schema := model.Schema[int]{
		Constraints: []model.Constraint[int]{constraint},
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
