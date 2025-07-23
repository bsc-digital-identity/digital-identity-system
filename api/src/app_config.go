package main

import (
	authschemas "api/src/auth_schemas"
	"api/src/identity"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func InitializeDev(db *gorm.DB) error {
	// Example: Insert admin if not exists
	admin_id, _ := uuid.NewRandom()
	admin := identity.SuperIdentity{
		IdentityId:   admin_id.String(),
		IdentityName: "admin",
	}

	result := db.FirstOrCreate(&admin)
	if result.Error != nil {
		log.Printf("Error inserting admin: %v", result.Error)
	}

	constraint := authschemas.Constraint[int]{
		Key:        "Age",
		Comparison: authschemas.GreaterThan,
		Value:      18,
	}

	schema := authschemas.Schema[int]{
		Constraints: []authschemas.Constraint[int]{constraint},
	}

	serialized_schema, err := json.Marshal(schema)
	if err != nil {
		log.Printf("Error when serializing: %s", err)
	}

	schema_id, _ := uuid.NewRandom()
	verifiable_schema := authschemas.VerifiableSchema{
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
