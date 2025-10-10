package integration

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Identity struct {
	Id           int `gorm:"primaryKey;autoIncrement"`
	IdentityId   string
	IdentityName string
	ParentId     *int
}

func TestGormPostgresConnection(t *testing.T) {
	dsn := "host=postgres user=api_user password=api_password dbname=digital_identity port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	t.Log("Successfully connected to database")

	if err := db.AutoMigrate(&Identity{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	t.Log("Migration succeeded")
}