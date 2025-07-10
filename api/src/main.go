package main

import (
	"api/src/database"
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from Go Docker multistage")
}

func main() {
	dbConnectionString := os.Getenv("DB_CONNECTION_STRING")
	database.ConnectToDatabase(dbConnectionString)
	/*
		id := uuid.New()
		adminIdentity := db.SuperIdentity{
			IdentityId:   id.String(),
			IdentityName: "admin",
		}
		rowId := database.Create(&adminIdentity)
		log.Printf("Created admin with %d Id", rowId)

		var suId db.SuperIdentity
		result := database.First(&suId, "identity_name = ?", "admin")
		log.Println(suId, result.Error)
	*/
	http.HandleFunc("/", handler)
	fmt.Println("server running at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
