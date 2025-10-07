package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Identity struct {
	Id           int    `gorm:"primaryKey;autoIncrement"`
	IdentityId   string
	IdentityName string
	ParentId     *int
}

func main() {
	dsn := "host=postgres user=api_user password=api_password dbname=digital_identity port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
	fmt.Println("Connected to database!")
	if err := db.AutoMigrate(&Identity{}); err != nil {
		panic(fmt.Sprintf("migration failed: %v", err))
	}
	fmt.Println("Migration succeeded!")
}
