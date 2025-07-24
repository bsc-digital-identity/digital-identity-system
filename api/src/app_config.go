package main

import (
	"api/src/model"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func InitializeDev(db *gorm.DB) error {
	// Example: Insert admin if not exists
	admin_id, _ := uuid.NewRandom()
	admin := model.Identity{
		IdentityId:   admin_id.String(),
		IdentityName: "admin",
	}

	result := db.FirstOrCreate(&admin)
	if result.Error != nil {
		log.Printf("Error inserting admin: %v", result.Error)
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
		log.Printf("Error when serializing: %s", err)
	}

	schema_id, _ := uuid.NewRandom()
	verifiable_schema := model.VerifiedSchema{
		SuperIdentityId: admin.Id,
		SchemaId:        schema_id.String(),
		Schema:          string(serialized_schema),
	}

	result = db.FirstOrCreate(&verifiable_schema)

	if result.Error != nil {
		log.Printf("Error when inserting schema: %s", result.Error)
	}

	log.Printf("Created admin user with id: %s", admin_id.String())
	log.Printf("Created schema with Age constraint: %s", schema_id.String())
	return nil
}
