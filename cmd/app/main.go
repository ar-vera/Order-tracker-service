package main

import (
	"Order-tracker-service/internal/db"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load("/Users/veraryabova/Desktop/Go/Order-tracker-service/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	pgDsn := os.Getenv("PG_DSN")

	fmt.Println(pgDsn)

	dataBase, err := sqlx.Connect("postgres", pgDsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dataBase.Close()

	db.RunMigrations(dataBase, "/Users/veraryabova/Desktop/Go/Order-tracker-service/migrations/")

}
